<template>
  <article class="train-item">
    <div class="wrap">
      <div class="header">
          候補 1/6
      </div>

      <div class="train">
      	<div class="departure">
      		<span class="time">11時53分 発</span>
      		<span class="station">新大阪</span>
      	</div>

      	<div class="name">
      		<span class="name">のぞみ 156 号</span>
      		<span class="type">N700系/全席禁煙</span>
      	</div>

      	<div class="arrival">
      		<span class="time">14時03分 着</span>
      		<span class="station">東京</span>
      	</div>
      </div>

      <div class="available">
        <p class="start">この候補を選択</p>
      </div>
    </div>
  </article>
</template>

<style scoped>

.train-item .header{
  margin:			0;
  padding:		0 0 0 20px;
  height:			30px;
  font-size:		12px;
  color:			#ffffff;
  line-height:		30px;
  font-weight:		normal;
  background: #1a65eb;
}

.train-item div.train {
    margin: 0;
    padding: 0 32px 0 43px;
    width: 565px;
}

.train-item div.departure,
.train-item div.arrival {
	margin:			0;
	padding:		0;
	height:			30px;
	font-size:		17px;
	color:			#ffffff;
	line-height:		30px;
}

.train-item .station {
  display: block;
  float: right;
  text-align: right;
}

.train-item div.departure {
	border-bottom:		1px solid #161B60;
}

.train-item div.arrival {
	border-top:		1px solid #161B60;
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
}


</style>
