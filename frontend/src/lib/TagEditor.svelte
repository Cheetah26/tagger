<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import type { Tag } from "$bindings/pkg/tagger";
  import Fuse from "fuse.js";
  import EditableTag from "./EditableTag.svelte";
  import { getTagString } from "./lib";
  import { appState } from "./state.svelte";

  let {
    tags,
    onAdd,
    onRemove,
  }: { tags: Tag[]; onAdd: (tag: Tag) => void; onRemove: (tag: Tag) => void } =
    $props();

  type Result = { tag: Tag; tagString: string };

  let search = $state("");

  let tagsWithStrings = $derived(
    appState.allTags
      ? Object.values(appState.allTags).map((t) => ({
          tag: t,
          tagString: getTagString(t),
        }))
      : [],
  );

  let fuse = $derived(
    new Fuse(tagsWithStrings, {
      keys: ["tag.name", "tagString"],
    }),
  );

  let results = $derived(fuse.search(search).slice(0, 3));

  async function newTagClicked() {
    if (!confirm(`Create new tag "${search}"?`)) {
      return;
    }
    const newTag = await TaggerService.AddTag(search);
    if (newTag) {
      appState.allTags = await TaggerService.GetAllTags();
      onAdd(newTag);
      search = "";
    }
  }

  function resultClicked(result: Result) {
    onAdd(result.tag);
    search = "";
  }
</script>

<p>Tags:</p>
{#each tags as tag}
  <EditableTag {tag} class="m-2 p-1 border"
    >{getTagString(tag)}<button
      onclick={() => onRemove(tag)}
      class="relative bottom-2 w-4 h-4 pl-1">x</button
    ></EditableTag
  >
{:else}
  <p>No tags</p>
{/each}

<div class="relative">
  <div class="flex flex-row align-middle">
    <input type="text" bind:value={search} class="w-full h-8" />
    <button onclick={newTagClicked}>+</button>
  </div>
  {#if search.length}
    <ul class="absolute bg-white border-2 border-black w-full">
      {#each results as result}
        <li>
          <button onclick={() => resultClicked(result.item)}
            >{result.item.tagString}</button
          >
        </li>
      {/each}
    </ul>
  {/if}
</div>
