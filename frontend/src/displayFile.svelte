<script lang="ts">
  import store from "./lib/store";
  import type { main } from "./lib/wailsjs/go/models";
  import TagFileChip from "./tagFileChip.svelte";

  export let file: main.File | undefined;

  // TODO: This might be leaving a trail of unused video players...
  let videoPlayer: HTMLMediaElement = document.createElement("video");

  // let dataBase64: string;
  // $: if (file) {
  //   dataBase64 = btoa(String.fromCharCode(...new Uint8Array(file.data)));
  // }

  let newTag = "";
  async function addTag() {
    if (!file) return;
    await store.tagFile(file, newTag);
    newTag = "";
  }

  async function removeFile() {
    if (file && confirm("Are you sure?")) {
      await store.removeFile(file);
    }
  }

  const imageFormats = [
    ".apng",
    ".avif",
    ".gif",
    ".jpg",
    ".jpeg",
    ".jfif",
    ".pjpeg",
    ".pjp",
    ".png",
    ".svg",
    ".webp",
  ];
  const videoFormats = [
    "webm",
    "mkv",
    "flv",
    ".vob",
    ".ogv",
    ".ogg",
    ".drc",
    ".gifv",
    ".mng",
    ".avi",
    ".mts",
    ".m2ts",
    ".TS",
    ".mov",
    ".qt",
    ".wmv",
    ".yuv",
    ".rm",
    ".rmvb",
    ".viv",
    ".asf",
    ".amv",
    ".mp4",
    ".m4p",
    ".m4v",
    ".svi",
    ".3gp",
    ".3g2",
    ".mxf",
    ".roq",
    ".nsv",
    ".flv",
    ".f4v",
    ".f4a",
    ".f4b",
  ];
</script>

{#if file === undefined}
  <p>No file selected</p>
{:else}
  <!-- Preview file -->
  {#if imageFormats.includes(file.filetype)}
    <img
      src={`data:image/${file.filetype};base64,${file.data}`}
      alt={file.hash}
    />
  {:else if videoFormats.includes(file.filetype) && videoPlayer.canPlayType(file.filetype) != ""}
    <video src={`data:video/${file.filetype};base64,${file.data}`}>
      <track kind="captions" />
    </video>
  {:else}
    <p class="bg-orange-500">Format {file.filetype} not supported</p>
  {/if}

  <p class="break-all">{file.hash.slice(0, 8)}</p>

  <p>Description:</p>
  {#if file.description}
    <p class="break-all">{file.description}</p>
    <p>HI</p>
  {/if}

  <!-- Tags -->
  <p>Tags:</p>
  {#if file.tags}
    {#each file.tags as tag}
      <TagFileChip {file} {tag} />
    {/each}
  {:else}
    <p>- No tags</p>
  {/if}
  <form
    on:submit|preventDefault={addTag}
    class="mt-2 flex flex-row align-middle"
  >
    <label for="new_tag">Add:</label>
    <input type="text" id="new_tag" class="w-full h-6" bind:value={newTag} />
  </form>

  <p class="mt-4">
    <button on:click={store.openCurrentFile}>Open</button>
    <button on:click={removeFile}>Remove</button>
  </p>
{/if}
