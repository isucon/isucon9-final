<template>
  <div class="loading" v-if="!items">
    読み込み中
  </div>

  <div class="trains" v-else>
    <section class="subcontent">
      <article class="condition">
        <div class="date">{{year}}年{{month}}月{{day}}日</div>
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
          グリーン車
        </div>
        <div class="th">
          <h3>指定席</h3>
          禁煙
        </div>
        <div class="td economy">
          <input type="radio" name="price" id="ek"/><label for="ek"></label>
          <div class="available">○</div>
          <div class="price">¥12,453</div>
        </div>
        <div class="td green">
          <input type="radio" name="price" id="gk"/><label for="gk"></label>
          <div class="available">○</div>
          <div class="price">¥12,453</div>
        </div>
        <div class="th">
          <h3>指定席</h3>
          禁煙（喫煙ルーム付近）
        </div>
        <div class="td economy">
          <input type="radio" name="price" id="es"/><label for="es"></label>
          <div class="available">○</div>
          <div class="price">¥12,453</div>
        </div>
        <div class="td green">
          <input type="radio" name="price" id="gs"/><label for="gs"></label>
          <div class="available">○</div>
          <div class="price">¥12,453</div>
        </div>
        <div class="th">
          <h3>自由席</h3>
        </div>
        <div class="td economy">
          <input type="radio" name="price" id="f"/><label for="f"></label>
          <div class="available">○</div>
          <div class="price">¥12,453</div>
        </div>
        <div class="td">
        </div>
      </div>


      <div class="seat">
        座席表を見る
      </div>

      <hr>

      <div class="position">
        <div>{{ position }}</div>
        <select v-model="position">
          <option>指定しない</option>
          <option>窓　側（普A／グA）</option>
          <option>中　央（普B）</option>
          <option>通路側（普C／グB）</option>
          <option>通路側（普D／グC）</option>
          <option>窓　側（普E／グD）</option>
        </select>
      </div>

      <div class="continue">
        予約を続ける
      </div>

      <div class="close" v-on:click="unselect()">
        閉じる
      </div>

    </div>

  </div>
</template>

<script>
import TrainItem from '@/components/Trains/TrainItem.vue'

export default {
  data: function() {
    return {
      year: null,
      month: null,
      day: null,
      from_station: null,
      to_station: null,
      adult: null,
      child: null,
      position: "指定しない",
      selectedItem: null,
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
        from_station: this.from_station,
        to_station: this.to_station,
        adult: this.adult,
        child: this.child,
      }
    }
  },
  methods: {
    unselect() {
      this.selectedItem = null;
    },
    select(item) {
      this.selectedItem = item;
    }
  },
  mounted() {
    this.year = this.$route.query.year
    this.month = this.$route.query.month
    this.day = this.$route.query.day
    this.adult = this.$route.query.adult
    this.child = this.$route.query.child
    this.from_station = this.$route.query.from_station
    this.to_station = this.$route.query.to_station

    this.items = [
      {
        "train_class": "のぞみ",
        "car_number": 95,
        "departure_at": new Date("2019-10-03 10:50:00"),
        "arrival_at": new Date("2019-10-03 12:32:00")
      },
      {
        "train_class": "こだま",
        "car_number": 50,
        "departure_at": new Date("2019-10-03 11:03:00"),
        "arrival_at": new Date("2019-10-03 12:52:00")
      }
    ]
  }
}
</script>

<style scoped>

div.trains {
  background: #18257F;

}

section.subcontent {
  width: 320px;
  float: left;
  background: #18257F;
  color: #ffffff;
}

section.trains {
  width: 640px;
  float: right;
  background: #82B1F9;
}

section.information {
  clear: both;
}

.condition {
  border-collapse: collapse;
  line-height: 1.1;
  padding: 10px;
}

.condition div {
  width: 100%;
  text-align: center;
  margin: 3px 0;
}

.condition .date {
  font-size: 30px;
}

.condition .station {
  font-size: 28px;
}

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
