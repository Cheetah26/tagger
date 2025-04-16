<script lang="ts">
  import { Pane, PaneGroup, PaneResizer } from "paneforge";
  import { appState } from "../state.svelte";
  import Details from "./Details.svelte";
  import Preview from "./Preview.svelte";

  let file = $derived(appState.selectedFile);
</script>

<PaneGroup direction="vertical" autoSaveId="display-file">
  <Pane defaultSize={1}>
    {#if file}
      {#key file.hash}
        <Preview {file}></Preview>
      {/key}
    {:else}
      <h1 class="w-full h-full pt-10 text-center">No file selected</h1>
    {/if}
  </Pane>
  <PaneResizer class="h-1 border"></PaneResizer>
  <Pane defaultSize={2} minSize={6}>
    {#if file}
      {#key file.id}
        <Details {file}></Details>
      {/key}
    {/if}
  </Pane>
</PaneGroup>
