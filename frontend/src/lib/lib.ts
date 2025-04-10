import type { Tag } from "$bindings/pkg/tagger";
import { appState } from "./state.svelte";

export function getTagString(tag: Tag): string {
  return tag.name + ' ' + tagStringHelper(tag);
}

function tagStringHelper(tag: Tag): string {
  if (!tag.parents || tag.parents.length == 0) {
    return "";
  }

  return `(${tag.parents.map(p => appState.allTags[p].name + tagStringHelper(appState.allTags[p])).join(", ")})`;
}