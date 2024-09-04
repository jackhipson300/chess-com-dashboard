<script>
    import DrawStatsBarCharts from "./components/DrawStatsBarCharts.svelte";
    import GameStatsBarCharts from "./components/GameStatsBarCharts.svelte";
    import LossStatsBarCharts from "./components/LossStatsBarCharts.svelte";
    import SearchBar from "./components/SearchBar.svelte";
    import WinStatsBarCharts from "./components/WinStatsBarCharts.svelte";
    import { fetchStats } from "./data/fetchStats";

  let username = ''
  let stats = undefined
  let error = undefined

  async function handleSearch(search) {
    try {
      stats = await fetchStats(search)
      console.log(stats)
    } catch(err) {
      error = err.message
      console.error(err)
    }
  }
</script>

<main>
  <div>
    <SearchBar onSearch={handleSearch} />
    <p>{username}</p>
    {#if error}<p>{error}</p>{/if}
    {#if stats}
      <div class="bar-chart-groups">
        <GameStatsBarCharts stats={stats} />
        <WinStatsBarCharts stats={stats} />
        <LossStatsBarCharts stats={stats} />
        <DrawStatsBarCharts stats={stats} />
      </div>
    {/if}
  </div>
</main>

<style>
  main {
    width: 80%;
    margin: auto;
  }

  .bar-chart-groups {
    display: flex;
    width: 100%;
    gap: 30px;
  }
</style>
