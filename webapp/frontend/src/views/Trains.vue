<template>
  <div class="trains">
    <section class="subcontent">
      <article class="condition">
        <div class="date">{{year}}年{{month}}月{{day}}日</div>
        <div class="station">{{ from_station }}→{{ to_station }}</div>
        <div class="person">おとな {{ adult }} 名 こども {{ child }} 名</div>
      </article>
    </section>

    <section class="trains">
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
      <TrainItem/>
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
  </div>
</template>

<script>
import TrainItem from '@/components/Trains/TrainItem.vue'

export default {
  data: function() {
    return {
      year: 2020,
      month: 10,
      day: 3,
      from_station: "東京",
      to_station: "新大阪",
      adult: 1,
      child: 2
    }
  },
  components: {TrainItem},
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


</style>
