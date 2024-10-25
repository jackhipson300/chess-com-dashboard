<script>
    import BarChartGroup from "./BarChartGroup.svelte";
    import StackedBarChart from "./StackedBarChart.svelte";

    export let stats

    const lossStatsColors = ["#DE4B50", "#CF2D32", "#913135", "#6F2628"]
</script>

<BarChartGroup header="Loss Results" labels={["Rapid", "Blitz", "Bullet"]}>
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
</BarChartGroup>