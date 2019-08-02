<template>
  <div>
    <section class="search">

      <article class="date">
        <select class="year" v-model="year">
          <option value="2020">2020年</option>
          <option value="2021">2021年</option>
        </select>
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

      <article class="station">
        <select class="from" v-model="from">
          <option>新大阪</option>
          <option>東京</option>
        </select>
        <span>→</span>
        <select class="to" v-model="to">
          <option>新大阪</option>
          <option>東京</option>
        </select>
      </article>

      <article class="person">
        <select class="adult" v-model="adult">
          <option value="0">おとな0名</option>
          <option value="1">おとな1名</option>
          <option value="2">おとな2名</option>
          <option value="3">おとな3名</option>
        </select>

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

    <section class="ui" style="float: none;">
      <article class="button-area" style="">
        <button>戻る</button>
        <button>予約を続ける</button>
      </article>
    </section>

  </div>
</template>

<script>
export default {
  name: 'reservation',
  components: {},
  data () {
    return {
      year: "2020",
      month: "1",
      day: "1",
      from: "",
      to: "",
      adult: "1",
      child: "0"
    }
  }
}
</script>

<style>

.search select {
    height: 50px;
    font-size: 24px;
    cursor: pointer;
    background: #0057D3;
    color: #ffffff;
    text-align: center;
}

.search {
  float: left;
}

.search .date .year {
    width: 214px;
}

.search .date .month {
    width: 213px;
}

.search .date .day {
    width: 213px;
}

.search .station .to,
.search .station .from {
    width: 300px;
}

.search .station span {
  vertical-align: top;
  padding: 0;
  margin: 0;
  height: 50px;
  width: 40px;
  display: inline-block;
  text-align: center;
  font-size: 24px;
  background: #0057D3;
  color: #ffffff;
}

.search .person .adult,
.search .person .child {
    width: 320px;
}




.subcontent {
  width: 317px;
  float: left;
}

.subcontent article {
    margin: 0;
    padding: 13px 19px;
    width: 279px;
    background:		#FFEAB4;
    height: 200px;
}

.subcontent article.notice h2{
	margin:			0 0 8px 0;
	padding:		0 0 8px 0;
	font-size:		18px;
	color:			#000000;
	line-height:		1.1;
	font-weight:		normal;
	border-bottom:		1px dashed #666666;
}

.subcontent article.notice p{
	margin:			0;
	padding:		0;
	font-size:		13px;
	color:			#666666;
	line-height:		1.4;
}


</style>
