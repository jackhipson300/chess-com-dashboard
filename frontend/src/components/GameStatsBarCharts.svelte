<script>
    import BarChartGroup from "./BarChartGroup.svelte";
    import StackedBarChart from "./StackedBarChart.svelte";

    export let stats

    const gameStatsColors = ["#81b64c", "#cacac8", "#cf2d32"]
</script>

<BarChartGroup header="Wins / Draws / Losses" labels={["Rapid", "Blitz", "Bullet"]}>
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
</BarChartGroup>