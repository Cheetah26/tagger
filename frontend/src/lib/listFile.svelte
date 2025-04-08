<script lang="ts">
  import { TaggerService } from "$bindings/index";
  import { File } from "../../bindings/github.com/cheetah26/tagger/pkg/tagger";
  import { appState } from "./state.svelte";

  export let file: File;

  function selectFile() {
    appState.selectedFile = file;
  }

  function openFile() {
    TaggerService.OpenFile(file);
  }
</script>

<button
  on:click={selectFile}
  on:dblclick={openFile}
  class="w-full text-left {appState.selectedFile?.id == file.id
    ? 'bg-green-300'
    : ''}">{file.id + "\t:\t" + file.hash.substring(0, 10)}</button
>
