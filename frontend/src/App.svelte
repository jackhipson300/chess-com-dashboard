<script>
  import axios from "axios";
  import SearchBar from "./components/SearchBar.svelte";

  let username = ''

  async function handleSearch(search) {
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
      }, 30 * 1000)

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
            setTimeout(poll, 500)
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

    result = await axios.get(`http://localhost:8090/gamestats?username=${search}`)
    console.log(result.data)
  }
</script>

<main>
  <div>
    <SearchBar onSearch={handleSearch} />
    <p>{username}</p>
  </div>
</main>

<style>
</style>
