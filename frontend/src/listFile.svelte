<script lang="ts">
  import store from "./lib/store";
  import { main } from "./lib/wailsjs/go/models";

  export let file: main.File;

  $: isCurrFile = $store.currentFile?.hash == file.hash;
</script>

<button
  on:click={() => store.selectFile(file)}
  on:dblclick={store.openCurrentFile}
  class={isCurrFile ? "bg-green-300" : ""}>{file.hash}</button
>
