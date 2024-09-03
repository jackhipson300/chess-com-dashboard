<script>
  export let search = ''
  export let onSearch = undefined

  function handleSubmit(event) {
    event.preventDefault()
    if(onSearch) {
      onSearch(search)
    }
  }
</script>

<form on:submit={handleSubmit}>
  <input
    type="text"
    bind:value={search}
    placeholder="Enter chess.com username"
  />
</form>

<style>

</style>