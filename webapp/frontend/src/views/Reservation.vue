<template>
  <div>
    <section class="search">
      <Date/>
      <Station/>
      <Person/>
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
import Date from '@/components/Reservation/Date.vue'
import Station from '@/components/Reservation/Station.vue'
import Person from '@/components/Reservation/Person.vue'

export default {
  name: 'reservation',
  components: {
    Date,
    Station,
    Person
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
