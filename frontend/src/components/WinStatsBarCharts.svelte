<script>
    import BarChartGroup from "./BarChartGroup.svelte";
    import StackedBarChart from "./StackedBarChart.svelte";

    export let stats

    const winStatsColors = ["#95DE4B", "#81B64C", "#4B6F26", "#344D1A"]
</script>

<BarChartGroup>
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
</BarChartGroup>