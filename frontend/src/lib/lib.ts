import type { main } from "./wailsjs/go/models";

export function getTagString(tag: main.Tag): string {
  let result = tag.name

  if (tag.parents == null || tag.parents == undefined) {
    return result
  }

  result += "("
  result += tag.parents.map(p => getTagString(p)).join(", ")
  result += ")"

  return result
}