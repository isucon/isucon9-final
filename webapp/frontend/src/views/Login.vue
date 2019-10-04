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
      <button type="button" v-on:click="onSubmit()">ログイン</button>
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
      })

    }
  }
}
</script>


<style scoped>
section {
  padding-top: 20px;
}

label {
  display: block;
  float: left;
  width: 150px;
  margin-left: 100px;
  color: #003163;
}
input {
  width: 200px;
  border-width: 1px;
  border-style: solid;
  border-color: #003163;
  font-size: 100%;
}

.form {
  float: left;
  width: 50%;
}

button {
    width: 300px;
    height: 65px;
    border-width: 0px;
    cursor: pointer;
    background: orange;
    margin-top: 15px;
    color: white;
    font-size: 25px;
    border-top-right-radius: 20px;
    border-bottom-right-radius: 20px;
}
</style>
