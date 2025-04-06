<script>
  import ListFile from "$lib/listFile.svelte";
  import store from "$lib/store";
  import DisplayFile from "$lib/displayFile.svelte";
  import TagListChip from "$lib/tagListChip.svelte";
  import { Pane, PaneGroup, PaneResizer } from "paneforge";

  let tagContainer = $state();
</script>

<main class="h-screen overflow-hidden p-2">
  <div class="flex flex-row border-b [&>*]:mr-2">
    <button onclick={store.open}>Open DB</button>
    <button onclick={store.importFiles}>Import</button>
    <button onclick={store.getUntaggedFiles}>Show Untagged files</button>
    <hr />
  </div>

  <PaneGroup direction="horizontal">
    <Pane defaultSize={1}>
      <div class="overflow-y-auto" bind:this={tagContainer}>
        <p>Tags: ({$store.tags ? Object.keys($store.tags).length : 0})</p>
        {#each Object.values($store.tags) as tag}
          <TagListChip {tag} contextMenuBounds={tagContainer}></TagListChip>
        {:else}
          <p>No tags</p>
        {/each}
      </div>
    </Pane>
    <PaneResizer class="w-1 m-1 border"></PaneResizer>
    <Pane defaultSize={1}>
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
    </Pane>
    <PaneResizer class="w-1 mx-1 border"></PaneResizer>
    <Pane defaultSize={2}>
      <DisplayFile />
    </Pane>
  </PaneGroup>
</main>
