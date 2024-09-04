<script>
  import { v4 } from "uuid";
  import * as d3 from 'd3';
  import { onMount } from "svelte";

  export let id = v4()
  export let width = 500;
  export let height = 30;
  export let data
  export let colors 

  function calculateTextDx(x, percent) {
    if(x <= 50) {
      return '.35em'
    }   

    const numDigits = `${percent.toFixed(0)}`.length
    const dx = `-${1 + ((numDigits - 1) * 0.65)}em`
    return dx
  }

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
        color: colors[i],
        percent: d.percent * 100
      }
    })

    const xScale = d3.scaleLinear()
      .domain([0, 100])
      .range([0, width])
    
    const bars = svg.selectAll('g.bar')
      .data(formattedData)
      .enter()
      .append('g')
      .attr('class', 'bar')

    bars
      .append('rect')
      .attr('x', (d) => xScale(d.x))
      .attr('width', (d) => xScale(Math.max(d.percent, 1)))
      .attr('height', height)
      .attr('fill', (_, i) => colors[i])

    bars
      .append('text')
      .attr('x', (d) => d.x > 50 ? xScale(d.x) + xScale(d.percent) : xScale(d.x))
      .attr('y', height / 2)
      .attr('dx', (d) => calculateTextDx(d.x, d.percent))
      .attr('dy', '.35em')
      .text((d) => d.percent > 1 ? d.percent.toFixed(0) : '')
      .attr('fill', 'black')
  })
</script>

<div>
  <svg id={id} />
</div>

<style>
  div {
    margin-top: 5px;
  }
</style>