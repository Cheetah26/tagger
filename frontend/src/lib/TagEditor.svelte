<script lang="ts">
  import type { Tag } from "$bindings/pkg/tagger";
  import EditableTag from "./EditableTag.svelte";
  import { getTagString } from "./lib";
  import { appState } from "./state.svelte";
  import { TaggerService } from "$bindings/index";
  import Fuse from "fuse.js";

  let {
    tags,
    onAdd,
    onRemove,
  }: {
    tags: Tag[];
    onAdd: (tag: Tag) => void | Promise<void>;
    onRemove: (tag: Tag) => void;
  } = $props();

  let search = $state("");
  let searchFocused = $state(true);

  let selectedItem = $state(0);

  let fuse = $derived.by(() => {
    let tagsWithStrings = Object.values(appState.allTags).map((tag) => ({
      tag,
      text: getTagString(tag),
    }));

    return new Fuse(tagsWithStrings, {
      keys: ["tag.name", "text"],
    });
  });

  let searchOptions = $derived(
    fuse
      .search(search)
      .slice(0, 3)
      .map((r) => r.item),
  );

  async function addTag(tag: Tag) {
    onAdd(tag);
    search = "";
  }

  async function addNewTag() {
    if (search.length <= 0) return;
    const newTag = await TaggerService.AddTag(search);
    if (newTag) {
      appState.allTags = await TaggerService.GetAllTags();
      addTag(newTag);
    }
  }

  function keydown(e: KeyboardEvent) {
    switch (e.code) {
      case "ArrowUp":
        selectedItem =
          selectedItem <= 0 ? searchOptions.length : selectedItem - 1;
        e.preventDefault();
        return;
      case "ArrowDown":
        selectedItem =
          selectedItem >= searchOptions.length ? 0 : selectedItem + 1;
        e.preventDefault();
        return;
      case "Enter":
        if (selectedItem < searchOptions.length) {
          addTag(searchOptions[selectedItem].tag);
        } else {
          addNewTag();
        }
        return;
    }
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

<div>
  <input
    type="text"
    bind:value={search}
    bind:focused={searchFocused}
    onkeydown={keydown}
    class="w-full h-8"
    placeholder="add tag"
  />
  {#if searchFocused}
    <div class="relative">
      <ul
        class="absolute top-0 left-0 w-full z-10 max-h-64 flex flex-col overflow-y-scroll bg-white border"
      >
        {#each searchOptions as opt, i}
          {@const selected = selectedItem == i}
          <li>
            <button
              onmousedown={(e) => {
                e.preventDefault();
                addTag(opt.tag);
              }}
              class="w-full h-full py-2 px-4 text-left hover:bg-gray-300"
              class:bg-gray-300={selected}>{opt.text}</button
            >
          </li>
        {/each}
        {#if search.length > 0}
          <li>
            <button
              onmousedown={(e) => {
                e.preventDefault();
                addNewTag();
              }}
              class="w-full h-full py-2 px-4 text-left hover:bg-gray-300"
              class:bg-gray-300={selectedItem >= searchOptions.length}
              >Create new tag: "{search}"</button
            >
          </li>
        {/if}
      </ul>
    </div>
  {/if}
</div>
