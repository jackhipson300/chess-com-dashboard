<script>
  export let header;
  export let labels = [];
</script>

<div class="outer">
  <div class="inner">
    {#if labels.length > 0}
      <div class="labels">
        {#each labels as label}
          <p>{label}</p>
        {/each}
      </div>
    {/if}
    <div>
      <p class="header">{header}</p>
      <slot />
    </div>
  </div>
</div>

<style>
  .outer {
    flex: 1;
  }
  .inner {
    display: flex;
  }
  .header {
    margin-bottom: 0.5rem;
  }
  .labels {
    margin-top: 2.25rem;
    margin-right: 0.5rem;
    text-align: right;
  }
</style>