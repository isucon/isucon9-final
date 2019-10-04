<template>
  <div class="loading" v-if="!items">
    読み込み中
  </div>

  <div class="trains" v-else>
    <section class="subcontent">
      <article class="condition">
        <div class="date">{{year}}年{{month}}月{{day}}日</div>
        <div class="time">{{hour}}時{{minute}}時 頃</div>
        <div class="station">{{ from_station }}→{{ to_station }}</div>
        <div class="person">おとな {{ adult }} 名 こども {{ child }} 名</div>
      </article>
    </section>

    <section class="trains">
      <TrainItem
        v-for="(item, index) in items"
        v-bind:item="item"
        v-bind:itemCount="itemCount"
        v-bind:itemIndex="index+1"
        v-bind:condition="condition"
        v-bind:select="select"
      />
    </section>

    <section class="information">
      <article>
        <h2>ご案内</h2>
        <p style="font-size: 11px;">
        ■空席表示について<br/>
        　「○」・・・空席あり<br/>
        　「▲」・・・残りわずか<br/>
        　「×」・・・満席<br/>
        　「－」・・・設定なし<br/>
        空席表示は混雑時間帯において、表示されない場合もございますのでご了承ください。<br/>
        ■発車時刻間際の列車は表示されません。
        </p>
      </article>
    </section>


    <div class="cover" v-if="selectedItem"></div>
    <div class="popup" v-if="selectedItem">

      <div class="prices">
        <div class="thead th"></div>
        <div class="thead td economy">
          普通車
        </div>
        <div class="thead td green">
          プレミアム車
        </div>
        <div class="th">
          <h3>指定席</h3>
          禁煙
        </div>
        <div class="td economy">
          <input type="radio" name="price" id="ek" v-bind:disabled="selectedItem.seat_availability.reserved == '×'" v-on:click="selectSeatClass('reserved', false)"/><label for="ek"></label>
          <div class="available">{{ selectedItem.seat_availability.reserved }}</div>
          <div class="price">¥{{ selectedItem.seat_fare.reserved }}</div>
        </div>
        <div class="td green">
          <input type="radio" name="price" id="gk" v-bind:disabled="selectedItem.seat_availability.premium == '×'" v-on:click="selectSeatClass('premium', false)"/><label for="gk"></label>
          <div class="available">{{ selectedItem.seat_availability.premium }}</div>
          <div class="price">¥{{ selectedItem.seat_fare.premium }}</div>
        </div>
        <div class="th">
          <h3>指定席</h3>
          禁煙（喫煙ルーム付近）
        </div>
        <div class="td economy">
          <input type="radio" name="price" id="es" v-bind:disabled="selectedItem.seat_availability.reserved_smoke == '×'" v-on:click="selectSeatClass('reserved', true)"/><label for="es"></label>
          <div class="available">{{ selectedItem.seat_availability.reserved_smoke }}</div>
          <div class="price">¥{{ selectedItem.seat_fare.reserved }}</div>
        </div>
        <div class="td green">
          <input type="radio" name="price" id="gs" v-bind:disabled="selectedItem.seat_availability.premium_smoke == '×'" v-on:click="selectSeatClass('premium', true)"/><label for="gs"></label>
          <div class="available">{{ selectedItem.seat_availability.premium_smoke }}</div>
          <div class="price">¥{{ selectedItem.seat_fare.premium }}</div>
        </div>
        <div class="th">
          <h3>自由席</h3>
        </div>
        <div class="td economy">
          <input type="radio" name="price" id="f" v-bind:disabled="selectedItem.seat_availability.non_reserved == '×'" v-on:click="selectSeatClass('non-reserved', false)"/><label for="f"></label>
          <div class="available">{{ selectedItem.seat_availability.non_reserved }}</div>
          <div class="price">¥{{ selectedItem.seat_fare.non_reserved }}</div>
        </div>
        <div class="td">
        </div>
      </div>


      <div class="seat" v-on:click="selectSeat()" v-bind:class="{ disabled: !canSelectSeat }">
        座席表を見る
      </div>

      <div class="position">
        <div>{{ position_display }}</div>
        <select v-model="position" v-bind:class="{ disabled: !columnChoices }">
          <option
            v-for="(choice, index) in columnChoices"
            v-bind:value="choice.value"
          >
            {{ choice.name }}
          </option>
        </select>
      </div>

      <div class="continue" v-on:click="reserve()" v-bind:class="{ disabled: !seat_class }">
        予約を続ける
      </div>

      <div class="close" v-on:click="unselect()">
        閉じる
      </div>

    </div>

  </div>
</template>

<script>
import Router from '@/router.js'
import TrainItem from '@/components/Trains/TrainItem.vue'
import { apiService } from '../services/api.js'

export default {
  data: function() {
    return {
      year: null,
      month: null,
      day: null,
      hour: null,
      minute: null,
      train_class: "",
      from_station: "",
      to_station: "",
      adult: null,
      child: null,
      position: "",
      selectedItem: null,
      seat_class: "",
      is_smoking_seat: false,
      items: null,
    }
  },
  components: {TrainItem},
  computed: {
    itemCount () {
      return this.items.length;
    },
    condition () {
      return {
        year: this.year,
        month: this.month,
        day: this.day,
        hour: this.hour,
        minute: this.minute,
        train_class: this.train_class,
        from_station: this.from_station,
        to_station: this.to_station,
        adult: this.adult,
        child: this.child,
      }
    },
    canSelectSeat () {
      if (this.seat_class == "" || this.seat_class == "non-reserved") {
        return false
      }
      return true
    },
    columnChoices () {
      if (this.seat_class == "premium") {
        return [
          {name: "指定しない", value:""},
          {name: "窓　側 (A)", value:"A"},
          {name: "通路側 (B)", value:"B"},
          {name: "通路側 (C)", value:"C"},
          {name: "窓　側 (D)", value:"D"},
        ]
      }
      if (this.seat_class == "reserved") {
        return [
          {name: "指定しない", value:""},
          {name: "窓　側 (A)", value:"A"},
          {name: "中　央 (B)", value:"B"},
          {name: "通路側 (C)", value:"C"},
          {name: "通路側 (D)", value:"D"},
          {name: "窓　側 (E)", value:"E"},
        ]
      }
      return []
    },
    position_display () {
      if (this.position == "") {
        return "指定しない"
      }

      if (this.seat_class == "premium") {
        var m = {
          A: "窓　側 (A)",
          B: "通路側 (B)",
          C: "通路側 (C)",
          D: "窓　側 (D)",
        }
        return m[this.position]
      }
      if (this.seat_class == "reserved") {
        var m = {
          A: "窓　側 (A)",
          B: "中　央 (B)",
          C: "通路側 (C)",
          D: "通路側 (D)",
          E: "窓　側 (E)",
        }
        return m[this.position]
      }
      return ""
    }
  },
  methods: {
    unselect() {
      this.selectedItem = null;
    },
    select(item) {
      this.selectedItem = item;
    },
    search() {
      apiService.getTrains(this.condition).then((res) => {
        console.log(res)
        var items = []

        res.forEach(function(value){
          value["departure"] = value["departure"]
          value["arrival"] = value["arrival"]
          value["departure_time"] = new Date("2000-01-01 " + value["departure_time"])
          value["arrival_time"] = new Date("2000-01-01 " + value["arrival_time"])
          items.push(value)
        });

        this.items = items
      })
    },
    selectSeatClass(seat_class, is_smoking_seat) {
      this.seat_class = seat_class
      this.is_smoking_seat = is_smoking_seat
    },
    selectSeat() {
      var query = {
        year: this.year,
        month: this.month,
        day: this.day,
        train_class: this.selectedItem.train_class,
        train_name: this.selectedItem.train_name,
        car_number: 4,
        from_station: this.from_station,
        to_station: this.to_station,
        adult: this.adult,
        child: this.child,
        seat_class: this.seat_class
      }
      if(this.seat_class!=""){
        Router.push({ path: '/reservation/seats', query: query})
      }
    },
    reserve() {
      if (this.seat_class == "") {
        return
      }
      var condition = {
        year: this.year,
        month: this.month,
        day: this.day,
        train_class: this.selectedItem.train_class,
        train_name: this.selectedItem.train_name,
        car_number: 0,
        from_station: this.from_station,
        to_station: this.to_station,
        adult: this.adult,
        child: this.child,
        seat_class: this.seat_class,
        is_smoking_seat: this.is_smoking_seat,
        column: this.position,
        seats: [],
      }

      apiService.reserve(condition).then((res) => {
        var query = {
          reservation_id: res.reservation_id,
        }
        Router.push({ path: '/reservation/payment', query: query})
      })
    }
  },
  mounted() {
    this.year = this.$route.query.year
    this.month = this.$route.query.month
    this.day = this.$route.query.day
    this.hour = this.$route.query.hour
    this.minute = this.$route.query.minute
    this.train_class = this.$route.query.train_class
    this.adult = parseInt(this.$route.query.adult)
    this.child = parseInt(this.$route.query.child)
    this.from_station = this.$route.query.from_station
    this.to_station = this.$route.query.to_station

    return this.search();
  }
}
</script>

<style scoped>

.cover {
  position: fixed;
  width: 100%;
  height: 100%;
  background: #000000;
  top: 0px;
  left: 0px;
  opacity: 0.3;
}

.popup {
  width: 460px;
  padding: 30px 0;
  background: #ffffff;
  position: absolute;
  left: 460px;
  top: 0px;
}

.popup .prices {
  margin-left: 20px;
}

.popup .prices .thead.th {
  background: #ffffff;
}

.popup .prices .thead.th,
.popup .prices .thead.td {
  height: 50px;
  text-align: center;
  line-height: 50px;
  border-style: solid;
  border-color: #ffffff;
  margin-bottom: 10px;
  padding: 0;
}

.popup .prices .thead.td.economy {
  background: #1C60EC;
}

.popup .prices .thead.td.green {
  background: #75AC33;
}

.popup .prices .th,
.popup .prices .td {
  float: left;
  margin: 0;
  height: 70px;
  padding: 0;
}

.popup .prices h3 {
  margin: 0;
}

.popup .prices .th {
  width: 227px;
  font-size: 12px;
  line-height: 1.2;
  color: #1D1F86;
  border-width: 1px 9px 1px 0;
  border-style: solid;
  border-color: #ffffff;
  background: #B9E3F8;
}

.popup .prices .td {
    width: 93px;
    font-size: 14px;
    line-height: 1.4;
    color: #ffffff;
    border-width: 1px 1px 1px 0;
    border-style: solid;
    border-color: #ffffff;
    text-align: center;
    height: 60px;
    padding-top: 10px;
}

.popup .prices .td.economy {
  background: #6E98F8;
}

.popup .prices .td.green {
  background: #98C33F;
}


/* ラベルのスタイル　*/
.popup .prices label {
	padding: 0;			/* ラベルの位置 */
	font-size:		10px;
	line-height:		20px;
	display:		inline-block;
	cursor:			pointer;
	position:		absolute;
  margin-top: -8px;
  margin-left: -45px;
}

/* ボックスのスタイル */
.popup .prices label:before {
	content:		'';
	width:			20px;			/* ボックスの横幅 */
	height:			20px;			/* ボックスの縦幅 */
	display:		inline-block;
	position:		absolute;
	left:			0;
	background-color:	#fff;
	box-shadow:		inset 1px 2px 3px 0px #000;
	border-radius:		6px 6px 6px 6px;
}
/* 元のチェックボックスを表示しない */
.popup .prices input[type=radio] {
	display:		none;
}
/* チェックした時のスタイル */
.popup .prices input[type=radio]:checked + label:before {
	content:		'\2713';		/* チェックの文字 */
	font-size:		20px;			/* チェックのサイズ */
	color:			#fff;			/* チェックの色 */
	background-color:	#06f;			/* チェックした時の色 */
}

/* チェックした時のスタイル */
.popup .prices input[type=radio]:disabled + label:before {
	background-color:	#555;			/* チェックした時の色 */
}

.popup .seat {
  margin-top: 10px;
  width: 424px;
  height: 60px;
  line-height: 60px;
  margin-left: 20px;
  text-align: center;
  float: left;
  background: #faae36;
  cursor: pointer;
}

.popup .seat.disabled {
  background: #aaaaaa;
  cursor: default;
}

.popup .continue.disabled {
  background: #aaaaaa;
  color: white;
  cursor: default;
}

.popup .position {
  float: left;
  position: relative;
  display: inline-block;
  margin-left: 20px;
  margin-top: 20px;
  height: 60px;
  line-height: 60px;
  cursor: pointer;
  vertical-align: top;
  color: #ffffff;
  font-size: 24px;
  background: #C9732B;
  width: 200px;
  font-size: 17px;
  text-align: center;
}

.popup .position select {
  position: absolute;
  left: 0px;
  top: 0px;
  height: 50px;
  font-size: 17px;
  opacity: 0;
  z-index: 5;
  cursor: pointer;
  width: 100%;
}


.popup .continue {
  margin-top: 20px;
  width: 204px;
  height: 60px;
  line-height: 60px;
  margin-left: 20px;
  text-align: center;
  float: left;
  background: pink;
  color: #ff0000;
  cursor: pointer;
}


.popup .close {
  margin-top: 20px;
  width: 100px;
  height: 40px;
  line-height: 40px;
  margin-left: 20px;
  text-align: center;
  float: left;
  background: #A0A0A0;
  cursor: pointer;
  color: #ffffff;
}


</style>
