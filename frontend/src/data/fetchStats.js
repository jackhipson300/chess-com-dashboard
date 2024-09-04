import axios from "axios";

const POLL_TIMEOUT = 30 * 1000
const POLL_INTERVAL = 1000

export async function fetchStats(search) {
  let result;
  try {
    result = await axios.post("http://localhost:8090/setup", { username: search })
  } catch(err) {
    if(!err.response || err.response.status !== 400) {
      throw err
    }
    console.error("Error requesting initial setup", err)
  }

  await new Promise(async (resolve, reject) => {
    if(!result?.data.id) {
      return reject(new Error("Error polling setup: no id in initial response"))
    }

    if(result.data.status === "Complete") {
      return resolve()
    }

    let running = true
    const timeout = setTimeout(() => {
      reject(new Error("Error polling setup: timeout"))
      running = false
    }, POLL_TIMEOUT)

    const poll = async () => {
      try {
        const pollRes = await axios.post("http://localhost:8090/setup", { username: search })
        if(!pollRes.data.status) {
          reject("Error polling setup: no status in response")
          return
        }

        if(pollRes.data.status === "Complete") {
          resolve()
          clearTimeout(timeout)
        } else {
          setTimeout(poll, POLL_INTERVAL)
        }
      } catch(err) {
        if(!err.response || err.response.status !== 400) {
          reject(err)
        }
        console.error("Error polling setup", err)
      }
    }

    poll()
  })

  const [gameStats, winStats, lossStats, drawStats] = await Promise.all([
    axios.get(`http://localhost:8090/gamestats?username=${search}`),
    axios.get(`http://localhost:8090/winstats?username=${search}`),
    axios.get(`http://localhost:8090/lossstats?username=${search}`),
    axios.get(`http://localhost:8090/drawstats?username=${search}`),
  ])

  return { 
    gameStats: gameStats.data, 
    winStats: winStats.data, 
    lossStats: lossStats.data, 
    drawStats: drawStats.data 
  }
}