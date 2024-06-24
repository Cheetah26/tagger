<script>
  import ListFile from "./listFile.svelte";
  import store from "./lib/store";
  import DisplayFile from "./displayFile.svelte";
  import TagChip from "./tagListChip.svelte";
</script>

<main class="grid grid-cols-[1fr_2fr_1fr] m-2">
  <div class="col-span-3">
    <button on:click={store.open}>Open DB</button>
    <button on:click={store.importFiles}>Import</button>
    <button on:click={store.getUntaggedFiles}>Show Untagged files</button>
    <hr />
  </div>

  <div>
    <p>Tags:</p>
    {#if $store.tags}
      {#each $store.tags as tag}
        <TagChip {tag}></TagChip>
      {/each}
    {:else}
      <p>No tags</p>
    {/if}
  </div>

  <div>
    <p>Files:</p>
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

  <div>
    <p>Selected:</p>
    <DisplayFile file={$store.currentFile} />
  </div>
</main>
