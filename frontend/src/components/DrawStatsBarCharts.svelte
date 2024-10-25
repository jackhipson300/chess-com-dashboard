<script>
    import BarChartGroup from "./BarChartGroup.svelte";
    import StackedBarChart from "./StackedBarChart.svelte";

  export let stats

  const drawStatsColors = ["#CACAC8", "#AAAAAA", "#888888", "#666666", "#444444", "#222222"]
</script>

<BarChartGroup header="Draw Results" labels={["Rapid", "Blitz", "Bullet"]}>
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
</BarChartGroup>