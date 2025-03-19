<script lang="ts">
  import type { Tag } from "../../bindings/github.com/cheetah26/tagger/pkg/tagger";
  import TagChip from "./TagChip.svelte";
  import store from "./store";
  import { getTagString } from "./lib";

  export let tags: Tag[] | undefined;
  export let onAdd: (tag: Tag) => void;
  export let onRemove: (tag: Tag) => void;

  type Result = { tag: Tag; tagString: string };

  let search = "";
  let results: Result[] = [];

  $: tagsWithStrings = $store.tags
    ? Object.values($store.tags).map((t) => ({
        tag: t,
        tagString: getTagString(t),
      }))
    : [];

  $: results = tagsWithStrings
    .filter((ts) => ts.tagString.includes(search))
    .slice(0, 3);

  async function newTagClicked() {
    if (!confirm('Add new tag "' + search + '"?')) {
      return;
    }
    const newTag = await store.addTag(search);
    if (newTag) {
      onAdd(newTag);
    }
    search = "";
  }

  function resultClicked(result: Result) {
    onAdd(result.tag);
    search = "";
  }
</script>

<p>Tags:</p>
{#if tags}
  {#each tags as tag}
    <TagChip {tag} cancelAction={() => onRemove(tag)}></TagChip>
  {/each}
{:else}
  <p>- No tags</p>
{/if}

<div class="relative">
  <div class="flex flex-row align-middle">
    <input type="text" bind:value={search} class="w-full h-8" />
    <button on:click={newTagClicked}>+</button>
  </div>
  {#if search.length}
    <ul class="absolute bg-white border-2 border-black w-full">
      {#each results as result}
        <li>
          <button on:click={() => resultClicked(result)}
            >{result.tagString}</button
          >
        </li>
      {/each}
    </ul>
  {/if}
</div>
