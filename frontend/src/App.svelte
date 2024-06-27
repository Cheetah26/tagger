<script>
  import ListFile from "./listFile.svelte";
  import store from "./lib/store";
  import DisplayFile from "./displayFile.svelte";
  import TagListChip from "./tagListChip.svelte";
  import Modal from "./lib/Modal.svelte";

  let tagContainer;

  let showModal = false;
</script>

<main
  class="max-h-screen overflow-hidden p-2 grid grid-cols-[1fr_2fr_1fr] grid-rows-[auto_1fr]"
>
  <div class="col-span-3">
    <button on:click={store.open}>Open DB</button>
    <button on:click={store.importFiles}>Import</button>
    <button on:click={store.getUntaggedFiles}>Show Untagged files</button>
    <button on:click={() => (showModal = true)}>Show Modal</button>
    <Modal bind:open={showModal}></Modal>
    <hr />
  </div>

  <div class="overflow-scroll" bind:this={tagContainer}>
    <p>Tags: ({$store.tags ? $store.tags.length : 0})</p>
    {#if $store.tags}
      {#each $store.tags as tag}
        <TagListChip {tag} contextMenuBounds={tagContainer}></TagListChip>
      {/each}
    {:else}
      <p>No tags</p>
    {/if}
  </div>

  <div class="overflow-scroll">
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

  <div class="overflow-y-scroll">
    <p>Selected:</p>
    <DisplayFile file={$store.currentFile} />
  </div>
</main>
