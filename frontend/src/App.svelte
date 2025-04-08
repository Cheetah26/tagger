<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import DisplayFile from "$lib/displayFile.svelte";
  import ListFile from "$lib/listFile.svelte";
  import { appState } from "$lib/state.svelte";
  import TagTree from "$lib/TagTree.svelte";
  import { Pane, PaneGroup, PaneResizer } from "paneforge";

  let rootTags = $derived(
    Array.from(Object.values(appState.allTags)).filter(
      (t) => t.parents.length == 0,
    ),
  );

  async function open() {
    await TaggerService.Open();
    appState.currentFiles = await TaggerService.GetAllFiles();
    appState.allTags = await TaggerService.GetAllTags();
  }

  async function importDialog() {
    await TaggerService.ImportFilesDialog();
    appState.getFiles();
  }
</script>

<main
  class="h-screen w-screen flex flex-col overflow-clip p-2 select-none cursor-default"
>
  <div class="flex flex-row shrink border-b [&>*]:mr-2">
    <button onclick={open}>Open DB</button>
    <button onclick={importDialog}>Import</button>
    <!-- <button onclick={store.getUntaggedFiles}>Show Untagged files</button> -->
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
        <p>
          Files: ({appState.currentFiles ? appState.currentFiles.length : 0})
        </p>
        {#if appState.currentFiles}
          <ul>
            {#each appState.currentFiles as file}
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
