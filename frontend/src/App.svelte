<script>
  import SearchBar from "./components/SearchBar.svelte";
    import { fetchStats } from "./data/fetchStats";
    import StackedBarChart from "./components/StackedBarChart.svelte";

  let username = ''
  let stats = undefined
  let error = undefined

  const gameStatsColors = ["#81b64c", "#cacac8", "#cf2d32"]
  const winStatsColors = ["#95DE4B", "#81B64C", "#4B6F26", "#344D1A"]
  const lossStatsColors = ["#DE4B50", "#CF2D32", "#913135", "#6F2628"]
  const drawStatsColors = ["#CACAC8", "#AAAAAA", "#888888", "#666666", "#444444", "#222222"]

  async function handleSearch(search) {
    try {
      stats = await fetchStats(search)
      console.log(stats)
    } catch(err) {
      error = err.message
      console.error(err)
    }
  }

  handleSearch("thatpurpledoor")
</script>

<main>
  <div>
    <SearchBar onSearch={handleSearch} />
    <p>{username}</p>
    {#if error}<p>{error}</p>{/if}
    {#if stats}
    <StackedBarChart 
      id="gameStatsRapidBar"
      data={[
        { percent: stats.gameStats.rapid.wins / stats.gameStats.rapid.total },
        { percent: stats.gameStats.rapid.draws / stats.gameStats.rapid.total },
        { percent: stats.gameStats.rapid.losses / stats.gameStats.rapid.total },
      ]}
      colors={gameStatsColors}
    />
    <StackedBarChart 
      id="gameStatsBlitzBar"
      data={[
        { percent: stats.gameStats.blitz.wins / stats.gameStats.blitz.total },
        { percent: stats.gameStats.blitz.draws / stats.gameStats.blitz.total },
        { percent: stats.gameStats.blitz.losses / stats.gameStats.blitz.total },
      ]}
      colors={gameStatsColors}
    />
    <StackedBarChart 
      id="gameStatsBulletBar"
      data={[
        { percent: stats.gameStats.bullet.wins / stats.gameStats.bullet.total },
        { percent: stats.gameStats.bullet.draws / stats.gameStats.bullet.total },
        { percent: stats.gameStats.bullet.losses / stats.gameStats.bullet.total },
      ]}
      colors={gameStatsColors}
    />
    <StackedBarChart 
      id="winStatsRapidBar"
      data={[
        { percent: stats.winStats.rapid.checkmates / stats.winStats.rapid.total },
        { percent: stats.winStats.rapid.resigns / stats.winStats.rapid.total },
        { percent: stats.winStats.rapid.abandons / stats.winStats.rapid.total },
        { percent: stats.winStats.rapid.timeouts / stats.winStats.rapid.total },
      ]}
      colors={winStatsColors}
    />
    <StackedBarChart 
      id="winStatsBlitzBar"
      data={[
        { percent: stats.winStats.blitz.checkmates / stats.winStats.blitz.total },
        { percent: stats.winStats.blitz.resigns / stats.winStats.blitz.total },
        { percent: stats.winStats.blitz.abandons / stats.winStats.blitz.total },
        { percent: stats.winStats.blitz.timeouts / stats.winStats.blitz.total },
      ]}
      colors={winStatsColors}
    />
    <StackedBarChart 
      id="winStatsBulletBar"
      data={[
        { percent: stats.winStats.bullet.checkmates / stats.winStats.bullet.total },
        { percent: stats.winStats.bullet.resigns / stats.winStats.bullet.total },
        { percent: stats.winStats.bullet.abandons / stats.winStats.bullet.total },
        { percent: stats.winStats.bullet.timeouts / stats.winStats.bullet.total },
      ]}
      colors={winStatsColors}
    />
    <StackedBarChart 
      id="lossStatsRapidBar"
      data={[
        { percent: stats.lossStats.rapid.checkmates / stats.lossStats.rapid.total },
        { percent: stats.lossStats.rapid.resigns / stats.lossStats.rapid.total },
        { percent: stats.lossStats.rapid.abandons / stats.lossStats.rapid.total },
        { percent: stats.lossStats.rapid.timeouts / stats.lossStats.rapid.total },
      ]}
      colors={lossStatsColors}
    />
    <StackedBarChart 
      id="lossStatsBlitzBar"
      data={[
        { percent: stats.lossStats.blitz.checkmates / stats.lossStats.blitz.total },
        { percent: stats.lossStats.blitz.resigns / stats.lossStats.blitz.total },
        { percent: stats.lossStats.blitz.abandons / stats.lossStats.blitz.total },
        { percent: stats.lossStats.blitz.timeouts / stats.lossStats.blitz.total },
      ]}
      colors={lossStatsColors}
    />
    <StackedBarChart 
      id="lossStatsBulletBar"
      data={[
        { percent: stats.lossStats.bullet.checkmates / stats.lossStats.bullet.total },
        { percent: stats.lossStats.bullet.resigns / stats.lossStats.bullet.total },
        { percent: stats.lossStats.bullet.abandons / stats.lossStats.bullet.total },
        { percent: stats.lossStats.bullet.timeouts / stats.lossStats.bullet.total },
      ]}
      colors={lossStatsColors}
    />
    <StackedBarChart 
      id="drawStatsRapidBar"
      data={[
        { percent: stats.drawStats.rapid.repetitions / stats.drawStats.rapid.total },
        { percent: stats.drawStats.rapid.stalemates / stats.drawStats.rapid.total },
        { percent: stats.drawStats.rapid.timeoutVsInsufficients / stats.drawStats.rapid.total },
        { percent: stats.drawStats.rapid.insufficients / stats.drawStats.rapid.total },
        { percent: stats.drawStats.rapid.agrees / stats.drawStats.rapid.total },
        { percent: stats.drawStats.rapid.fiftyMoveRules / stats.drawStats.rapid.total },
      ]}
      colors={drawStatsColors}
    />
    <StackedBarChart 
      id="drawStatsBlitzBar"
      data={[
        { percent: stats.drawStats.blitz.repetitions / stats.drawStats.blitz.total },
        { percent: stats.drawStats.blitz.stalemates / stats.drawStats.blitz.total },
        { percent: stats.drawStats.blitz.timeoutVsInsufficients / stats.drawStats.blitz.total },
        { percent: stats.drawStats.blitz.insufficients / stats.drawStats.blitz.total },
        { percent: stats.drawStats.blitz.agrees / stats.drawStats.blitz.total },
        { percent: stats.drawStats.blitz.fiftyMoveRules / stats.drawStats.blitz.total },
      ]}
      colors={drawStatsColors}
    />
    <StackedBarChart 
      id="drawStatsBulletBar"
      data={[
        { percent: stats.drawStats.bullet.repetitions / stats.drawStats.bullet.total },
        { percent: stats.drawStats.bullet.stalemates / stats.drawStats.bullet.total },
        { percent: stats.drawStats.bullet.timeoutVsInsufficients / stats.drawStats.bullet.total },
        { percent: stats.drawStats.bullet.insufficients / stats.drawStats.bullet.total },
        { percent: stats.drawStats.bullet.agrees / stats.drawStats.bullet.total },
        { percent: stats.drawStats.bullet.fiftyMoveRules / stats.drawStats.bullet.total },
      ]}
      colors={drawStatsColors}
    />
    {/if}
  </div>
</main>

<style>
</style>
