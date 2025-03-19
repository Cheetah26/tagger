import { get } from "svelte/store";
import type { Tag } from "../../bindings/github.com/cheetah26/tagger/pkg/tagger";
import store from "./store";

export function getTagString(tag: Tag): string {
  let result = tag.name

  if (!tag.parents || tag.parents.length == 0) {
    return result
  }

  result += "("
  result += tag.parents.map(p => getTagString(get(store).tags[p])).join(", ")
  result += ")"

  return result
}