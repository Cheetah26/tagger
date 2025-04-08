import type { Tag } from "$bindings/pkg/tagger";
import { appState } from "./state.svelte";

export function getTagString(tag: Tag): string {
  let result = tag.name

  if (!tag.parents || tag.parents.length == 0) {
    return result
  }

  result += "("
  result += tag.parents.map(p => getTagString(appState.allTags[p])).join(", ")
  result += ")"

  return result
}