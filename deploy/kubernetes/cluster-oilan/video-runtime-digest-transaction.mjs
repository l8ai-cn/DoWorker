import {
  closeSync,
  copyFileSync,
  existsSync,
  fsyncSync,
  mkdirSync,
  openSync,
  readFileSync,
  renameSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { dirname, isAbsolute, join, relative, resolve } from "node:path";

const TRANSACTION_DIR = "deploy/kubernetes/cluster-oilan/.video-runtime-digest-transaction";
const LOCK_FILE = "deploy/kubernetes/cluster-oilan/.video-runtime-digest-update.lock";

export function withinDigestUpdateLock(root, apply) {
  const lockPath = join(root, LOCK_FILE);
  let locked = false;
  try {
    mkdirSync(lockPath);
    locked = true;
    writeFileSync(join(lockPath, "owner"), `${process.pid}\n`);
    syncFile(join(lockPath, "owner"));
    recoverInterruptedUpdate(root);
    return apply();
  } finally {
    if (locked) {
      rmSync(lockPath, { recursive: true });
    }
  }
}

export function writeRecoverably(root, updates) {
  const transactionPath = join(root, TRANSACTION_DIR);
  mkdirSync(transactionPath);
  const records = updates.map((update, index) => ({
    target: repoRelativePath(root, update.path),
    backup: `${index}.backup`,
  }));
  const staged = [];

  try {
    for (const [index, update] of updates.entries()) {
      const backupPath = join(transactionPath, records[index].backup);
      copyFileSync(update.path, backupPath);
      syncFile(backupPath);
    }
    writeFileSync(
      join(transactionPath, "manifest.json"),
      `${JSON.stringify({ version: 1, records }, null, 2)}\n`,
    );
    syncFile(join(transactionPath, "manifest.json"));

    for (const [index, update] of updates.entries()) {
      const stagedPath = join(dirname(update.path), `.video-runtime-${process.pid}-${index}.tmp`);
      const fd = openSync(stagedPath, "wx", statSync(update.path).mode);
      writeFileSync(fd, update.content);
      fsyncSync(fd);
      closeSync(fd);
      staged.push(stagedPath);
    }
    for (const [index, update] of updates.entries()) {
      renameSync(staged[index], update.path);
    }
    rmSync(transactionPath, { recursive: true });
  } catch (error) {
    restoreTransaction(root, transactionPath);
    throw error;
  } finally {
    for (const path of staged) {
      rmSync(path, { force: true });
    }
  }
}

function recoverInterruptedUpdate(root) {
  const transactionPath = join(root, TRANSACTION_DIR);
  if (!existsSync(transactionPath)) return;
  const manifestPath = join(transactionPath, "manifest.json");
  if (!existsSync(manifestPath)) {
    rmSync(transactionPath, { recursive: true });
    return;
  }
  restoreTransaction(root, transactionPath);
}

function restoreTransaction(root, transactionPath) {
  const manifestPath = join(transactionPath, "manifest.json");
  if (!existsSync(manifestPath)) {
    rmSync(transactionPath, { recursive: true, force: true });
    return;
  }
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  if (manifest.version !== 1 || !Array.isArray(manifest.records)) {
    throw new Error("invalid video runtime digest transaction manifest");
  }
  for (const record of manifest.records) {
    if (!/^\d+\.backup$/.test(record.backup)) {
      throw new Error("invalid digest transaction backup");
    }
    const target = resolveRepoPath(root, record.target);
    const backup = join(transactionPath, record.backup);
    copyFileSync(backup, target);
  }
  rmSync(transactionPath, { recursive: true });
}

function repoRelativePath(root, path) {
  const target = relative(root, path);
  if (target === "" || target.startsWith("..") || isAbsolute(target)) {
    throw new Error(`digest update target is outside repository: ${path}`);
  }
  return target;
}

function resolveRepoPath(root, path) {
  if (typeof path !== "string") {
    throw new Error("invalid digest transaction target");
  }
  return join(root, repoRelativePath(root, resolve(root, path)));
}

function syncFile(path) {
  const fd = openSync(path, "r");
  fsyncSync(fd);
  closeSync(fd);
}
