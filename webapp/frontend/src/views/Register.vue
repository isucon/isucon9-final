<template>
  <div>

  <form>

  <section class="register">

    <article class="form">
      <p>
        <label for="email">メールアドレス</label>
        <input type="email" id="email" size="" maxlength="100" placeholder="example@example.com" v-model="email">
      </p>
      <p>
        <label for="password">パスワード</label>
        <input type="password" id="password" size="" maxlength="100" placeholder="" v-model="password">
      </p>
    </article>
    <article class="button">
      <button type="button" v-on:click="onSubmit()">登録</button>
    </article>

  </section>

  </form>

  </div>

</template>

<script>
import Router from '@/router.js'
import { apiService } from '../services/api.js'

export default {
  name: 'register',
  components: {},
  data() {
    return {
      email: "",
      password: "",
    }
  },
  methods:{
    onSubmit() {
      apiService.register({
        email: this.email, password: this.password,
      }).then((res) => {
        Router.push({ path: '/login' })
      }).catch((error) => {
        console.log(error)
        alert(error.response.data.message)
      })

      return false
    }
  }
}
</script>
