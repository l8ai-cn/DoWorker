import type { SplitTreeNode, SplitTreeLeaf, SplitTreeSplit } from "./workspaceTypes";

export const generatePaneId = () => `pane-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;
export const generateNodeId = () => `node-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;

export function findLastLeaf(node: SplitTreeNode): SplitTreeLeaf | null {
  if (node.type === "leaf") return node;
  for (let i = node.children.length - 1; i >= 0; i--) {
    const found = findLastLeaf(node.children[i]);
    if (found) return found;
  }
  return null;
}

export function findLeafByPaneId(node: SplitTreeNode, paneId: string): SplitTreeLeaf | null {
  if (node.type === "leaf") return node.paneId === paneId ? node : null;
  for (const child of node.children) {
    const found = findLeafByPaneId(child, paneId);
    if (found) return found;
  }
  return null;
}

export function replaceNode(tree: SplitTreeNode, nodeId: string, replacement: SplitTreeNode): SplitTreeNode {
  if (tree.id === nodeId) return replacement;
  if (tree.type === "leaf") return tree;
  return { ...tree, children: tree.children.map((c) => replaceNode(c, nodeId, replacement)) };
}

export function findParentSplit(tree: SplitTreeNode, targetId: string): SplitTreeSplit | null {
  if (tree.type === "leaf") return null;
  for (const child of tree.children) {
    if (child.id === targetId) return tree;
    const found = findParentSplit(child, targetId);
    if (found) return found;
  }
  return null;
}

export function insertChildAt(
  tree: SplitTreeNode, splitId: string,
  child: SplitTreeNode, afterIndex: number, sizes: number[],
): SplitTreeNode {
  if (tree.type === "leaf") return tree;
  if (tree.id === splitId) {
    const newChildren = [...tree.children];
    newChildren.splice(afterIndex + 1, 0, child);
    return { ...tree, children: newChildren, sizes };
  }
  return { ...tree, children: tree.children.map((c) => insertChildAt(c, splitId, child, afterIndex, sizes)) };
}

export function removeLeaf(tree: SplitTreeNode, leafId: string): SplitTreeNode | null {
  if (tree.type === "leaf") return tree.id === leafId ? null : tree;

  const idx = tree.children.findIndex((c) => c.id === leafId);
  if (idx >= 0) {
    const remaining = tree.children.filter((_, i) => i !== idx);
    if (remaining.length === 0) return null;
    if (remaining.length === 1) return remaining[0]; // collapse
    const remainingSizes = tree.sizes.filter((_, i) => i !== idx);
    const total = remainingSizes.reduce((a, b) => a + b, 0);
    return { ...tree, children: remaining, sizes: remainingSizes.map((s) => (s / total) * 100) };
  }

  const newChildren: SplitTreeNode[] = [];
  const removedIndices: number[] = [];
  for (let i = 0; i < tree.children.length; i++) {
    const result = removeLeaf(tree.children[i], leafId);
    if (result) {
      newChildren.push(result);
    } else {
      removedIndices.push(i);
    }
  }
  if (newChildren.length === 0) return null;
  if (newChildren.length === 1) return newChildren[0]; // collapse

  const keptSizes = tree.sizes.filter((_, i) => !removedIndices.includes(i));
  const total = keptSizes.reduce((a, b) => a + b, 0);
  return { ...tree, children: newChildren, sizes: keptSizes.map((s) => (s / total) * 100) };
}

export function updateSizes(tree: SplitTreeNode, splitId: string, sizes: number[]): SplitTreeNode {
  if (tree.type === "leaf") return tree;
  if (tree.id === splitId) return { ...tree, sizes };
  return { ...tree, children: tree.children.map((c) => updateSizes(c, splitId, sizes)) };
}

export function appendPaneLeaf(
  tree: SplitTreeNode,
  newLeaf: SplitTreeLeaf,
  activePaneId: string | null,
): SplitTreeNode {
  const activeLeaf = activePaneId ? findLeafByPaneId(tree, activePaneId) : null;
  const parent = activeLeaf ? findParentSplit(tree, activeLeaf.id) : null;

  if (parent && activeLeaf) {
    const index = parent.children.findIndex((child) => child.id === activeLeaf.id);
    const evenSize = 100 / (parent.children.length + 1);
    const sizes = Array.from(
      { length: parent.children.length + 1 },
      () => evenSize,
    );
    return insertChildAt(tree, parent.id, newLeaf, index, sizes);
  }

  return {
    type: "split",
    id: generateNodeId(),
    direction: "horizontal",
    children: [tree, newLeaf],
    sizes: [50, 50],
  };
}

export function splitPaneLeaf(
  tree: SplitTreeNode,
  leaf: SplitTreeLeaf,
  newLeaf: SplitTreeLeaf,
  direction: "horizontal" | "vertical",
): SplitTreeNode {
  const parent = findParentSplit(tree, leaf.id);

  if (parent && parent.direction === direction) {
    const index = parent.children.findIndex((child) => child.id === leaf.id);
    const targetSize = parent.sizes[index];
    const sizes = [...parent.sizes];
    sizes[index] = targetSize / 2;
    sizes.splice(index + 1, 0, targetSize / 2);
    return insertChildAt(tree, parent.id, newLeaf, index, sizes);
  }

  return replaceNode(tree, leaf.id, {
    type: "split",
    id: generateNodeId(),
    direction,
    children: [{ ...leaf }, newLeaf],
    sizes: [50, 50],
  });
}
