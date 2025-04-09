<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import { Pane, PaneGroup, PaneResizer } from "paneforge";
  import { appState } from "./state.svelte";

  let tableWidth = $state(0);
  let idColWidth = $state(0);

  let idColPane = $state(undefined);
</script>

{#if appState.currentFiles}
  <table class="flex flex-col w-full h-full" bind:clientWidth={tableWidth}>
    <thead>
      <tr class="block w-ful border-b">
        <PaneGroup direction="horizontal" autoSaveId="list-files">
          <Pane defaultSize={1} class="min-w-fit" bind:id={idColPane}
            ><th class="block w-full" bind:clientWidth={idColWidth}>File Id</th
            ></Pane
          >
          <PaneResizer class="w-1 m-1 border"></PaneResizer>
          <Pane defaultSize={4} class="min-w-fit"
            ><th class="block w-full">Description</th></Pane
          >
        </PaneGroup>
      </tr>
    </thead>
    <tbody class="overflow-x-hidden overflow-y-scroll">
      {#each appState.currentFiles as file}
        {@const selected = appState.selectedFile?.id == file.id}
        <tr
          onclick={() => (appState.selectedFile = file)}
          ondblclick={() => TaggerService.OpenFile(file)}
          class="flex flex-row {selected && 'bg-green-300'}"
        >
          <td
            style="width: {idColWidth}px;"
            class="whitespace-nowrap overflow-hidden overflow-ellipsis"
            >{file.id}</td
          >
          <td
            style="width: {tableWidth - idColWidth}px;"
            class="whitespace-nowrap overflow-hidden overflow-ellipsis"
            >{file.description}</td
          >
        </tr>
      {/each}
    </tbody>
  </table>
{:else}
  <h1>No files in current selection</h1>
{/if}
