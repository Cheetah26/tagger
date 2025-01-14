<script lang="ts">
  export let open: boolean;
</script>

{#if open}
  <popover
    class="z-20 absolute top-0 left-0 w-screen h-screen bg-black/50 flex justify-center items-center"
  >
    <div class="relative w-80 min-h-20 bg-white rounded-md p-4">
      <button
        on:click={() => (open = false)}
        class="absolute top-2 right-2 w-6 h-6">X</button
      >
      <slot></slot>
    </div>
  </popover>
{/if}
