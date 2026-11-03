interface ClassificationTreeNode {
  id: string;
  parentId?: string;
  name: string;
  slug: string;
}

/** Preserve catalog ordering and ancestors when a search matches a descendant. */
export function classificationFamilies<T extends ClassificationTreeNode>(nodes: T[], search = "") {
  const byId = new Map(nodes.map(node => [node.id, node]));
  const term = search.trim().toLocaleLowerCase();
  const matches = (node: T) => `${node.name} ${node.slug}`.toLocaleLowerCase().includes(term);
  const families: { root: T; children: T[] }[] = [];
  const visited = new Set<string>();
  function descendants(node: T): T[] {
    if (visited.has(node.id)) return [];
    visited.add(node.id);
    return nodes.filter(child => child.parentId === node.id).flatMap(child => [child, ...descendants(child)]);
  }
  for (const root of nodes.filter(node => !node.parentId || !byId.has(node.parentId))) {
    const children = descendants(root);
    if (!term || matches(root)) families.push({ root, children });
    else {
      const found = children.filter(matches);
      if (found.length) families.push({ root, children: found });
    }
  }
  return families;
}

export function classificationPath<T extends ClassificationTreeNode>(node: T, nodes: T[]): string {
  const names = [node.name];
  const seen = new Set([node.id]);
  let parent = nodes.find(item => item.id === node.parentId);
  while (parent && !seen.has(parent.id)) {
    names.unshift(parent.name);
    seen.add(parent.id);
    parent = nodes.find(item => item.id === parent?.parentId);
  }
  return names.join(" / ");
}
