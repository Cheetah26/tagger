<script>
  import ListFile from "./lib/listFile.svelte";
  import store from "./lib/store";
  import DisplayFile from "./lib/displayFile.svelte";
  import TagListChip from "./lib/tagListChip.svelte";
  import { get } from "svelte/store";

  let tagContainer;
</script>

<main
  class="h-screen overflow-hidden p-2 grid grid-cols-[1fr_2fr_1fr] grid-rows-[auto_1fr]"
>
  <div class="col-span-3">
    <button on:click={store.open}>Open DB</button>
    <button on:click={store.importFiles}>Import</button>
    <button on:click={store.getUntaggedFiles}>Show Untagged files</button>
    <hr />
  </div>

  <div class="overflow-y-auto" bind:this={tagContainer}>
    <p>Tags: ({$store.tags ? Object.keys($store.tags).length : 0})</p>
    {#if $store.tags}
      {#each Object.values($store.tags) as tag}
        <TagListChip {tag} contextMenuBounds={tagContainer}></TagListChip>
      {/each}
    {:else}
      <p>No tags</p>
    {/if}
  </div>

  <div class="overflow-y-scroll">
    <p>Files: ({$store.files ? $store.files.length : 0})</p>
    {#if $store.files}
      <ul>
        {#each $store.files as file}
          <li>
            <ListFile {file}></ListFile>
          </li>
        {/each}
      </ul>
    {:else}
      <h1>No files in current selection</h1>
    {/if}
  </div>

  <div class="overflow-y-auto">
    <p>Selected:</p>
    <DisplayFile />
  </div>
</main>
