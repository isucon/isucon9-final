<template>
  <div>

  <form>

  <section class="register">
    <input v-model="card_number"/>
    <input v-model="cvv"/>
    <input v-model="expiry_date"/>
    <button type="button" v-on:click="onSubmit()">登録</button>
  </section>

  </form>

  </div>

</template>

<script>
import Router from '@/router.js'
import { apiService } from '../services/api.js'

export default {
  name: 'test',
  components: {},
  data() {
    return {
      card_number: "00000000",
      cvv: "000",
      expiry_date: "12/24",
    }
  },
  methods:{
    onSubmit() {

      var data = {
        card_number: this.card_number,
        cvv: this.cvv,
        expiry_date: this.expiry_date,
      }

      apiService.tokenizeCard(data).then((res) => {
        console.log("OK")
        console.log(res)
      }).catch((error) => {
        console.log("ERROR")
        console.log(error)
        console.log(error.data)
      })

      return false
    }
  }
}
</script>
