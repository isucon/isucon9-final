<template>
  <div id="app">
    <header>
      <div class="welcome">
        <p class="logo"><a></a></p>
        <p class="welcome" v-if="user"><span>{{ user.email }} 様</span><span style="color:#ff0000"></span></p>
      </div>

      <div class="title">
        <ul>
          <li class="b1"><router-link to="/"><a>メニュー</a></router-link></li>
          <li class="b2" v-if="!user"><a href="#" v-on:click="login()">ログイン</a></li>
          <li class="b2" v-if="user"><a href="#" v-on:click="logout()">ログアウト</a></li>
          <li class="b3"><a href="#" v-on:click="register()">ユーザー登録</a></li>
        </ul>
      </div>

    </header>
    <router-view/>
  </div>
</template>

<script>
import Router from '@/router.js'
import { apiService } from './services/api.js'

export default {
  data() {
    return {
      user: null,
    }
  },
  methods:{
    register() {
      Router.push({ path: '/register'})
    },
    login() {
      Router.push({ path: '/login'})
    },
    logout() {
      apiService.logout().then((res) => {
        this.user = null;
        Router.push({ path: '/login' })
      })
    },
    getAuth() {
      apiService.getAuth().then((res) => {
        this.user = res
      })
    }
  },
  mounted() {
    this.getAuth()
  }
}
</script>

<style>
#app {
  font-family: "Hiragino Kaku Gothic Pro", "ヒラギノ角ゴ Pro W3", "メイリオ", Meiryo, "ＭＳ Ｐゴシック", Helvetica, Arial, Verdana, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  margin:      0 auto;
  padding:    0;
  width:      960px;
}

/* header */

header{
  margin:      0;
  padding:    0;
}

header div.welcome{
  margin:      0;
  padding:    0;
  height:      50px;
  background:    #DAEFFB;
  position:    relative;
}

header div.welcome p.logo{
  margin:      0;
  padding:    0;
  width:      129px;
  height:      50px;
}

header div.welcome p.logo a{
  display:    block;
  margin:      0;
  padding:    50px 0 0 0;
  width:      300px;
  height:      0;
  overflow:    hidden;
  background:    url(assets/logo.svg) no-repeat 5px 5px;
}


header div.welcome p.welcome{
  margin:      0;
  padding:    0;
  font-size:    14px;
  color:      #003264;
  line-height:    1;
  position:    absolute;
  right:      20px;
  bottom:      8px;
}

header div.welcome p.welcome span{
  font-size:    17px;
}

header div.title{
  margin:      0;
  padding:    0;
  height:      48px;
  background:    #003163;
  position:    relative;
}

header div.title h1{
  margin:      0;
  padding:    0;
  height:      48px;
  position:    absolute;
  left:      100px;
  top:      0;
  font-size:    16px;
  color:      #ffffff;
  line-height:    48px;
  font-weight:    normal;
}

header div.title ul{
  margin:      0;
  padding:    0;
  list-style:    none;
}

header div.title ul li.b1{
  margin:      0;
  padding:    0;
  position:    absolute;
  left:      22px;
  top:      9px;
}

header div.title ul li.b1 a{
  display:    inline-block;
  /display:    inline;
  /zoom:      1;
  margin:      0;
  padding:    0;
  width:      65px;
  height:      30px;
  background:    #ffffff;
  border-radius:    8px;
  font-size:    12px;
  color:      #000000;
  line-height:    30px;
  text-align:    center;
  text-decoration:  none;
}

header div.title ul li.b1 a:hover{
  background:    #CCCCCC;
}

header div.title ul li.b2{
  margin:      0;
  padding:    0;
  position:    absolute;
  right:      130px;
  top:      9px;
}

header div.title ul li.b2 a{
  width:      90px;
  height:      30px;
}

header div.title ul li.b3{
  margin:      0;
  padding:    0;
  position:    absolute;
  right:      22px;
  top:      9px;
}

header div.title ul li.b3 a{
  width:      100px;
  height:      30px;
}

header div.title ul li.b2 a ,
header div.title ul li.b3 a{
  display:    inline-block;
  /display:    inline;
  /zoom:      1;
  margin:      0;
  padding:    0;
  background:    #CCD6E0;
  border-radius:    8px;
  font-size:    12px;
  color:      #000000;
  line-height:    30px;
  text-align:    center;
  text-decoration:  none;
}

header div.title ul li.b2 a:hover ,
header div.title ul li.b3 a:hover {
  background:    #B7C5D2;
}


# Home


.home article {
  display: block;
  opacity: 1;
}

.home article:hover {
  opacity: 0.8;
}
.home article p {
  margin:      0;
  padding:    0;
  display:    block;
  padding:    77px 0 0 135px;
  width:      505px;
  height:      133px;
  font-size:    42px;
  color:      #ffffff;
  line-height:    1.1;
  font-style:    normal;
  text-decoration:  none;
  letter-spacing:    2px;
}

.home article a {
  color: #ffffff;
  text-decoration: none;
}

.home article.reserve{
  width:      960px;
  height:      210px;
  background:  #1C1F84;
}

.home article.inquiry{
  width:      960px;
  height:      210px;
  background:  #00A6ED;
}


.information {
  margin:      0;
  padding:    13px 19px;
  background:    #FFEAB4;
  width:      922px;
}

.information article h2{
  margin:      0 0 8px 0;
  padding:    0 0 8px 0;
  font-size:    18px;
  color:      #000000;
  line-height:    1.1;
  font-weight:    normal;
  border-bottom:    2px dotted #666666;
}

.information article h2 span{
  position:    absolute;
  left:      -5000px;
  top:      -5000px;
}

.information article p{
  margin:      0;
  padding:    0;
  font-size:    13px;
  color:      #666666;
  line-height:    1.4;
}

div.trains {
  background: #18257F;
  color: white;
}

div.trains section.subcontent {
  width: 320px;
  float: left;
  background: #18257F;
  color: #ffffff;
}

div.trains section.trains {
  width: 640px;
  float: right;
  background: #82B1F9;
}

div.trains section.price {
  width: 640px;
  float: right;
  background: #1E39D5;
}

div.trains section.information {
  clear: both;
}

div.trains .condition {
  border-collapse: collapse;
  line-height: 1.1;
  padding: 10px;
}

div.trains .condition div {
  width: 100%;
  text-align: center;
  margin: 3px 0;
}

div.trains .condition .date {
  font-size: 30px;
}

div.trains .condition .time {
  font-size: 30px;
}

div.trains .condition .station {
  font-size: 28px;
}


.train-item .header{
  margin:      0;
  padding:    0 0 0 20px;
  height:      30px;
  font-size:    12px;
  color:      #ffffff;
  line-height:    30px;
  font-weight:    normal;
  background: #1a65eb;
}

.train-item div.train {
    margin: 0;
    padding: 0 32px 0 43px;
    width: 565px;
}

.train-item div.departure,
.train-item div.arrival {
  margin:      0;
  padding:    0;
  height:      30px;
  font-size:    17px;
  color:      #ffffff;
  line-height:    30px;
}

.train-item .station {
  display: block;
  float: right;
  text-align: right;
}

.train-item div.departure {
  border-bottom:    1px solid #161B60;
}

.train-item div.arrival {
  border-top:    1px solid #161B60;
}

.train-item div.name span {
  display: block;
  margin: 0;
  padding: 0 0 0 30px;
  color: #172A87;
  font-weight: normal;
}

.train-item div.name span.name {
  font-size: 25px;
  line-height: 27px;
}

.train-item .available {
  color: #ffffff;
  margin-left: auto;
  width: 215px;
  height: 42px;
  line-height: 42px;
  padding-left: 20px;
  background: #e59836;
  cursor: pointer;
}

.train-item .seats h3 {
  color: #18257F;
  font-size: 30px;
}

.train-item .seats p {
  text-align: right;
  color: #18257F;
  font-size: 30px;
}

.price-item p {
  margin: 0;
}

.price-item {
  font-size: 20px;
}
.price-item .wrap {
  width: 200px;
  margin-left: auto;
  margin-right: auto;
}


</style>
