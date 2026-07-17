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

const DEFAULT_NAMESPACE = "video-runtime";

export function withinDigestUpdateLock(root, apply, namespace = DEFAULT_NAMESPACE) {
  const lockPath = join(root, lockFile(namespace));
  let locked = false;
  try {
    acquireDigestUpdateLock(lockPath);
    locked = true;
    writeFileSync(join(lockPath, "owner"), `${process.pid}\n`);
    syncFile(join(lockPath, "owner"));
    recoverInterruptedUpdate(root, namespace);
    return apply();
  } finally {
    if (locked) {
      rmSync(lockPath, { recursive: true });
    }
  }
}

function acquireDigestUpdateLock(lockPath) {
  try {
    mkdirSync(lockPath);
    return;
  } catch (error) {
    if (error.code !== "EEXIST" || !lockOwnerExited(lockPath)) {
      throw error;
    }
  }
  rmSync(lockPath, { recursive: true });
  mkdirSync(lockPath);
}

function lockOwnerExited(lockPath) {
  let owner;
  try {
    owner = Number.parseInt(readFileSync(join(lockPath, "owner"), "utf8"), 10);
  } catch {
    return false;
  }
  if (!Number.isSafeInteger(owner) || owner <= 0) return false;
  try {
    process.kill(owner, 0);
    return false;
  } catch (error) {
    return error.code === "ESRCH";
  }
}

export function writeRecoverably(root, updates, namespace = DEFAULT_NAMESPACE) {
  const transactionPath = join(root, transactionDir(namespace));
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
      const stagedPath = join(dirname(update.path), `.${namespace}-${process.pid}-${index}.tmp`);
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

function recoverInterruptedUpdate(root, namespace) {
  const transactionPath = join(root, transactionDir(namespace));
  if (!existsSync(transactionPath)) return;
  const manifestPath = join(transactionPath, "manifest.json");
  if (!existsSync(manifestPath)) {
    rmSync(transactionPath, { recursive: true });
    return;
  }
  restoreTransaction(root, transactionPath);
}

function transactionDir(namespace) {
  requireNamespace(namespace);
  return `deploy/kubernetes/cluster-oilan/.${namespace}-digest-transaction`;
}

function lockFile(namespace) {
  requireNamespace(namespace);
  return `deploy/kubernetes/cluster-oilan/.${namespace}-digest-update.lock`;
}

function requireNamespace(namespace) {
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(namespace)) {
    throw new Error(`invalid digest update namespace: ${namespace}`);
  }
}

function restoreTransaction(root, transactionPath) {
  const manifestPath = join(transactionPath, "manifest.json");
  if (!existsSync(manifestPath)) {
    rmSync(transactionPath, { recursive: true, force: true });
    return;
  }
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  if (manifest.version !== 1 || !Array.isArray(manifest.records)) {
    throw new Error("invalid runtime digest transaction manifest");
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
