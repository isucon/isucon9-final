<template>
  <div>
    <div class="search">
      <section class="search">

        <article class="year">
          <div>{{ year }}年</div>
          <select class="year" v-model="year">
            <option value="2020">2020年</option>
            <option value="2021">2021年</option>
          </select>
        </article>
        <article class="month">
          <div>{{ month }}月</div>
          <select class="month" v-model="month">
            <option value="1">1月</option>
            <option value="2">2月</option>
            <option value="3">3月</option>
            <option value="4">4月</option>
            <option value="5">5月</option>
            <option value="6">6月</option>
            <option value="7">7月</option>
            <option value="8">8月</option>
            <option value="9">9月</option>
            <option value="10">10月</option>
            <option value="11">11月</option>
            <option value="12">12月</option>
          </select>
        </article>
        <article class="day">
          <div>{{ day }}日</div>
          <select class="day" v-model="day">
            <option value="1">1日</option>
            <option value="2">2日</option>
            <option value="3">3日</option>
            <option value="4">4日</option>
            <option value="5">5日</option>
            <option value="6">6日</option>
            <option value="7">7日</option>
            <option value="8">8日</option>
            <option value="9">9日</option>
            <option value="10">10日</option>
            <option value="11">11日</option>
            <option value="12">12日</option>
            <option value="13">13日</option>
            <option value="14">14日</option>
            <option value="15">15日</option>
            <option value="16">16日</option>
            <option value="17">17日</option>
            <option value="18">18日</option>
            <option value="19">19日</option>
            <option value="20">20日</option>
            <option value="21">21日</option>
            <option value="22">22日</option>
            <option value="23">23日</option>
            <option value="24">24日</option>
            <option value="25">25日</option>
            <option value="26">26日</option>
            <option value="27">27日</option>
            <option value="28">28日</option>
            <option value="29">29日</option>
            <option value="30">30日</option>
            <option value="31">31日</option>
          </select>
        </article>


        <article class="hour">
          <div>{{ hour }}時</div>
          <select class="hour" v-model="hour">
            <option value="6">06時</option>
            <option value="7">07時</option>
            <option value="8">08時</option>
            <option value="9">09時</option>
            <option value="10">10時</option>
            <option value="11">11時</option>
            <option value="12">12時</option>
            <option value="13">13時</option>
            <option value="14">14時</option>
            <option value="15">15時</option>
            <option value="16">16時</option>
            <option value="17">17時</option>
            <option value="18">18時</option>
            <option value="19">19時</option>
            <option value="20">20時</option>
            <option value="21">21時</option>
            <option value="22">22時</option>
            <option value="23">23時</option>
          </select>
        </article>
        <article class="minute">
          <div>{{ minute }}分</div>
          <select class="minute" v-model="minute">
            <option value="0">00分</option>
            <option value="10">10分</option>
            <option value="20">20分</option>
            <option value="30">30分</option>
            <option value="40">40分</option>
            <option value="50">50分</option>
          </select>
        </article>
        <article class="minute-after">
          <div>頃</div>
        </article>



        <article class="train_class">
          <div>{{ train_class }}</div>
          <select class="train_class" v-model="train_class">
            <option value="全て">全て</option>
            <option value="最速">最速</option>
            <option value="中間">中間</option>
            <option value="遅いやつ">遅いやつ</option>
          </select>
        </article>

        <article class="from">
          <div>{{ from_station.name }}</div>
          <select class="from" v-model="from_station_id">
            <option v-for="station in usableStations" v-bind:key="station.id" :value="station.id">
              {{ station.name }}
            </option>
          </select>
        </article>
        <article class="arrow">
          <div>→</div>
        </article>
        <article class="to">
          <div>{{ to_station.name }}</div>
          <select class="to" v-model="to_station_id">
            <option v-for="station in usableStations" v-bind:key="station.id" :value="station.id">
              {{ station.name }}
            </option>
          </select>
        </article>

        <article class="adult">
          <div>おとな{{ adult }}名</div>
          <select class="adult" v-model="adult">
            <option value="0">おとな0名</option>
            <option value="1">おとな1名</option>
            <option value="2">おとな2名</option>
            <option value="3">おとな3名</option>
          </select>
        </article>
        <article class="child">
          <div>こども{{ child }}名</div>
          <select class="child" v-model="child">
            <option value="0">こども0名</option>
            <option value="1">こども1名</option>
            <option value="2">こども2名</option>
            <option value="3">こども3名</option>
          </select>
        </article>


      </section>
      <section class="subcontent">
        <article class="notice">
          <h2>ご案内</h2>
          <p style="font-size: 11px;">
            ■本サービスの商品は、乗継駅等で途中下車することはできません。<br>
          </p>
        </article>
      </section>

      <div style="clear: both;"></div>

    </div>

    <section class="ui" style="float: none;">
      <article class="button-area" style="">
        <button v-on:click="search()">予約を続ける</button>
      </article>
    </section>
  </div>
</template>

<script>
import Router from '@/router.js'
import { apiService } from '../services/api.js'

export default {
  name: 'search',
  components: {},
  data () {
    return {
      year: 2020,
      month: 1,
      day: 1,
      hour: 6,
      minute: 0,
      train_class: "全て",
      from_station_id: 0,
      to_station_id: 0,
      adult: "1",
      child: "0",
      stations: []
    }
  },
  computed: {
    from_station() {
      return apiService.getStation(this.from_station_id)
    },
    to_station() {
      return apiService.getStation(this.to_station_id)
    },
    usableStations() {
      var ret = this.stations

      if (this.train_class == "最速"){
        ret = this.stations.filter(station => {
          return station.is_stop_express;
        })
      }

      if (this.train_class == "中間"){
        ret = this.stations.filter(station => {
          return station.is_stop_semi_express;
        })
      }

      if (this.train_class == "遅いやつ"){
        ret = this.stations.filter(station => {
          return station.is_stop_local;
        })
      }

      return ret
    },
    train_class_query() {
      if(this.train_class=="全て") {
        return ""
      }
      return this.train_class
    }
  },
  methods: {
    loadStations() {
      apiService.getStations().then((res) => {
        console.log(res)
        this.stations = res
        this.from_station_id = res[0].id
        this.to_station_id = res[res.length-1].id
      })
    },
    search() {
      var query = {
        year: this.year,
        month: this.month,
        day: this.day,
        hour: this.hour,
        minute: this.minute,
        train_class: this.train_class_query,
        from_station: this.from_station.name,
        to_station: this.to_station.name,
        adult: this.adult,
        child: this.child
      }
      Router.push({ path: '/reservation/trains', query: query})
    }
  },
  mounted(){
    this.loadStations()
  }
}
</script>

<style scoped>

div.search {
  background: #FFEAB4;
}

section.search {
  width: 640px;
  margin: 0;
  float: left;
}

.search article {
  position: relative;
  display: inline-block;
  height: 50px;
  cursor: pointer;
  vertical-align: top;
  color: #ffffff;
  font-size: 24px;
}

.search article div {
  padding-top: 10px;
  line-height: 1.2;
  text-align: center;
}

.search select {
  position: absolute;
  left: 0px;
  top: 0px;
  height: 50px;
  font-size: 24px;
  opacity: 0;
  z-index: 5;
  cursor: pointer;
  width: 100%;
}

article.year {
  width: 214px;
  height: 80px;
  font-size: 40px;
  background: #1C1F84;
}


article.month {
  width: 213px;
  height: 80px;
  font-size: 40px;
  background: #1C1F84;
}

article.day {
  width: 213px;
  height: 80px;
  font-size: 40px;
  background: #1C1F84;
}

article.hour {
  width: 250px;
  background: #0057D3;
}


article.minute {
  width: 250px;
  background: #0057D3;
}


article.minute-after {
  width: 140px;
  background: #0057D3;
}

article.train_class {
  background: #1C1F84;
  width: 640px;
}

article.from ,
article.to {
  height: 60px;
  font-size: 30px;
  background: #0057D3;
  width: 300px;
}

article.arrow {
  height: 60px;
  background: #0057D3;
  width: 40px;
}


article.adult ,
article.child {
  background: #1C1F84;
  width: 320px;
}


.subcontent {
  width: 317px;
  float: left;
}

.subcontent article {
    margin: 0;
    padding: 13px 19px;
    width: 270px;
    background:    #FFEAB4;
    height: 200px;
}

.subcontent article.notice h2{
  margin:      0 0 8px 0;
  padding:    0 0 8px 0;
  font-size:    18px;
  color:      #000000;
  line-height:    1.1;
  font-weight:    normal;
  border-bottom:    1px dashed #666666;
}

.subcontent article.notice p{
  margin:      0;
  padding:    0;
  font-size:    13px;
  color:      #666666;
  line-height:    1.4;
}

.button-area {
  text-align:center;
}
button {
  width: 300px;
  height: 50px;
  margin: 30px;
  background: pink;
  color: red;
  font-size: 20px;
  border-width: 1px;
  border-color: #999999;
  border-top-left-radius: 20px;
  border-bottom-left-radius: 20px;
  border-top-right-radius: 20px;
  border-bottom-right-radius: 20px;
  font-weight: bold;
}


</style>
