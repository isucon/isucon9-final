<template>
  <div>

  <form v-on:submit="onSubmit()">

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
      <button type="submit">ログイン</button>
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
      apiService.login({
        email: this.email, password: this.password,
      }).then((res) => {
        Router.push({ path: '/' })
      }).catch((error) => {
        console.log(error)
        alert(error.response.data.message)
      })

      return false
    }
  }
}
</script>
