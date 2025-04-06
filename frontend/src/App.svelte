<script>
  import ListFile from "$lib/listFile.svelte";
  import store from "$lib/store";
  import DisplayFile from "$lib/displayFile.svelte";
  import TagListChip from "$lib/tagListChip.svelte";
  import { Pane, PaneGroup, PaneResizer } from "paneforge";
  import TagTree from "$lib/TagTree.svelte";

  let rootTags = $derived(
    Array.from(Object.values($store.tags)).filter((t) => t.parents.length == 0),
  );
</script>

<main class="h-screen overflow-hidden p-2 select-none cursor-default">
  <div class="flex flex-row border-b [&>*]:mr-2">
    <button onclick={store.open}>Open DB</button>
    <button onclick={store.importFiles}>Import</button>
    <button onclick={store.getUntaggedFiles}>Show Untagged files</button>
    <hr />
  </div>

  <PaneGroup direction="horizontal">
    <Pane defaultSize={1} class="min-w-fit">
      <div class="h-full overflow-y-scroll pr-6">
        <ul>
          {#each rootTags as tag}
            <TagTree {tag}></TagTree>
          {/each}
        </ul>
      </div>
    </Pane>
    <PaneResizer class="w-1 m-1 border"></PaneResizer>
    <Pane defaultSize={1}>
      <div class="h-full overflow-y-scroll">
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
