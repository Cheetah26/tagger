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
    let possibleResults = Object.values(appState.allTags)
      .filter((t) => !tags.some((existing) => existing.id == t.id))
      .map((tag) => ({
        tag,
        text: getTagString(tag),
      }));

    return new Fuse(possibleResults, {
      keys: ["tag.name", "text"],
    });
  });

  let results = $derived(fuse.search(search, { limit: 3 }).map((r) => r.item));

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
        selectedItem = selectedItem >= results.length ? 0 : selectedItem + 1;
        e.preventDefault();
        return;
      case "ArrowDown":
        selectedItem = selectedItem <= 0 ? results.length : selectedItem - 1;
        e.preventDefault();
        return;
      case "Enter":
        if (selectedItem < results.length) {
          addTag(results[selectedItem].tag);
        } else {
          addNewTag();
        }
        return;
    }
  }
</script>

<div>
  <p>Tags:</p>
  <div>
    {#if searchFocused && search.length > 0}
      <div class="relative">
        <ul
          class="absolute bottom-0 left-0 w-full z-10 max-h-64 flex flex-col-reverse overflow-y-scroll bg-white border"
        >
          {#each results as opt, i}
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
                class:bg-gray-300={selectedItem >= results.length}
                >Create new tag: "{search}"</button
              >
            </li>
          {/if}
        </ul>
      </div>
    {/if}
    <input
      type="text"
      bind:value={search}
      bind:focused={searchFocused}
      onkeydown={keydown}
      class="w-full h-8"
      placeholder="add tag"
    />
  </div>

  {#each tags as tag}
    <EditableTag {tag} class="m-1 p-1 border"
      >{getTagString(tag)}<button
        onclick={() => onRemove(tag)}
        class="relative bottom-2 w-4 h-4 pl-1">x</button
      ></EditableTag
    >
  {:else}
    <p>No tags</p>
  {/each}
</div>
