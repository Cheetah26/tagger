<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import FilePane from "$lib/File/FilePane.svelte";
  import ListFiles from "$lib/FileList.svelte";
  import { appState } from "$lib/state.svelte";
  import TagTree from "$lib/TagTree.svelte";
  import { Pane, PaneGroup, PaneResizer } from "paneforge";

  let rootTagIds = $derived(
    appState.tagIdsOrdered.filter(
      (id) => appState.allTags[id].parents.length == 0,
    ),
  );

  async function open() {
    await TaggerService.Open();

    appState.currentFiles = await TaggerService.GetAllFiles();
    appState.selectedFile = undefined;

    appState.allTags = await TaggerService.GetAllTags();
    appState.tagIdsOrdered = await TaggerService.GetTagIdsOrdered(
      appState.tagOrdering,
    );
    appState.selectedTags = [];
  }

  async function importDialog() {
    await TaggerService.ImportFilesDialog();
    appState.getFiles();
  }

  async function clear() {
    appState.showUntagged = false;
    appState.selectedTags = [];
    appState.allTags = await TaggerService.GetAllTags();
    appState.currentFiles = await TaggerService.GetAllFiles();
  }

  async function toggleUntagged() {
    appState.showUntagged = !appState.showUntagged;
    if (appState.showUntagged) {
      appState.selectedTags = [];
    }
    appState.getFiles();
  }
  $effect(() => {
    if (appState.selectedTags.length > 0) {
      appState.showUntagged = false;
      appState.getFiles();
    }
  });
</script>

<main
  class="h-screen w-screen flex flex-col overflow-clip select-none cursor-default"
>
  <div class="flex flex-row shrink border-b [&>*]:mr-2">
    <button onclick={open}>Open DB</button>
    <button onclick={importDialog}>Import</button>
    <hr />
  </div>

  <PaneGroup direction="horizontal" autoSaveId="app">
    <Pane defaultSize={1} minSize={15} class="p-1">
      <div class="h-full overflow-y-scroll pr-5">
        <button class="w-full" onclick={clear}>Clear Filters</button>

        <div class="flex flex-row items-center border-b">
          <label for="untagged" class="flex-grow">Untagged</label>
          <input
            type="checkbox"
            id="untagged"
            checked={appState.showUntagged}
            onclick={toggleUntagged}
            class="m-1"
          />
        </div>

        <ul>
          {#each rootTagIds as id}
            <TagTree tag={appState.allTags[id]}></TagTree>
          {/each}
        </ul>
      </div>
    </Pane>
    <PaneResizer class="w-1 border"></PaneResizer>
    <Pane defaultSize={1} minSize={15} class="p-1">
      <ListFiles />
    </Pane>
    <PaneResizer class="w-1 border"></PaneResizer>
    <Pane defaultSize={2} minSize={25}>
      <FilePane />
    </Pane>
  </PaneGroup>
</main>
