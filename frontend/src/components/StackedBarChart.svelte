<script>
  import { v4 } from "uuid";
  import * as d3 from 'd3';
  import { onMount } from "svelte";

  export let id = v4()
  export let width = 500;
  export let height = 30;
  export let data
  export let colors 

  onMount(() => {
    const svg = d3.select(`#${id}`)
      .attr('width', width)
      .attr('height', height)
    
    let xPos = 0
    const formattedData = data.map((d, i) => {
      const temp = xPos
      xPos += d.percent * 100
      return {
        ...d,
        x: temp,
        color: colors[i]
      }
    })

    const xScale = d3.scaleLinear()
      .domain([0, 100])
      .range([0, width])
    
    svg.selectAll('rect')
      .data(formattedData)
      .enter()
      .append('rect')
      .attr('x', (d) => xScale(d.x))
      .attr('width', (d) => xScale(Math.max(d.percent * 100, 1)))
      .attr('height', height)
      .attr('fill', (_, i) => colors[i])
  })
</script>

<div>
  <svg id={id} />
</div>