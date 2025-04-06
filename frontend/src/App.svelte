<script>
  import ListFile from "$lib/listFile.svelte";
  import store from "$lib/store";
  import DisplayFile from "$lib/displayFile.svelte";
  import TagListChip from "$lib/tagListChip.svelte";
  import ResizablePaneGroup from "$lib/components/ui/resizable/resizable-pane-group.svelte";
  import { ResizablePane } from "$lib/components/ui/resizable";
  import ResizableHandle from "$lib/components/ui/resizable/resizable-handle.svelte";

  let tagContainer;
</script>

<main class="h-screen overflow-hidden p-2">
  <div class="flex flex-row justify-between border-b">
    <button on:click={store.open}>Open DB</button>
    <button on:click={store.importFiles}>Import</button>
    <button on:click={store.getUntaggedFiles}>Show Untagged files</button>
    <hr />
  </div>
  <ResizablePaneGroup direction="horizontal">
    <ResizablePane>
      <div class="h-full w-full overflow-y-scroll" bind:this={tagContainer}>
        <p>Tags: ({$store.tags ? Object.keys($store.tags).length : 0})</p>
        {#if $store.tags}
          {#each Object.values($store.tags) as tag}
            <TagListChip {tag} contextMenuBounds={tagContainer}></TagListChip>
          {/each}
        {:else}
          <p>No tags</p>
        {/if}
      </div>
    </ResizablePane>
    <ResizableHandle></ResizableHandle>
    <ResizablePane>
      <div class="h-full w-full overflow-y-scroll">
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
    </ResizablePane>
    <ResizableHandle></ResizableHandle>
    <ResizablePane>
      <div class="h-full w-full overflow-y-auto">
        <p>Selected:</p>
        <DisplayFile />
      </div>
    </ResizablePane>
  </ResizablePaneGroup>
</main>
