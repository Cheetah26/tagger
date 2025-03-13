<script lang="ts">
  import store from "./lib/store";
  import { File } from "../bindings/github.com/cheetah26/tagger/pkg/tagger";

  export let file: File;

  $: highlight = $store.currentFile?.id == file.id ? "bg-green-300" : "";
</script>

<button
  on:click={() => store.selectFile(file)}
  on:dblclick={store.openCurrentFile}
  class="w-full text-left {highlight}"
  >{file.id + "\t:\t" + file.hash.substring(0, 10)}</button
>
