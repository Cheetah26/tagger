<script lang="ts">
  import store from "./lib/store";
  import type { main } from "./lib/wailsjs/go/models";

  export let file: main.File;
  export let tag: main.Tag;

  const BaseStyle = "border-2 border-black m-1 p-1 text-sm ";
</script>

<span class={BaseStyle + "inline-block"}
  >{tag.name}
  <button on:click={() => store.untagFile(file, tag)}>X</button></span
>
