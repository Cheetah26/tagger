<script lang="ts">
  import store from "./lib/store";
  import { tagger } from "./lib/wailsjs/go/models";

  export let file: tagger.File;

  $: highlight = $store.currentFile?.id == file.id ? "bg-green-300" : "";
</script>

<button
  on:click={() => store.selectFile(file)}
  on:dblclick={store.openCurrentFile}
  class="w-full text-left {highlight}"
  >{file.id + "\t:\t" + file.hash.substring(0, 10)}</button
>
