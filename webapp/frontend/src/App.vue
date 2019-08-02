<template>
  <div id="app">
    <header>
      <div class="welcome">
    		<p class="logo"><a></a></p>
    		<p class="welcome2"><span>最終ログイン日時：2019/08/02 16:32</span></p>
    		<p class="welcome"><span>こたまご様</span><span style="color:#ff0000"></span></p>
      </div>

      <div class="title">
    		<h1>メニュー</h1>
    		<ul>
      		<li class="b1"><router-link to="/"><a>メニュー</a></router-link></li>
      		<li class="b2"><a href="#">ログアウト</a></li>
    		</ul>
    	</div>

    </header>
    <router-view/>
  </div>
</template>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  margin:			0 auto;
	padding:		0;
	width:			960px;
}

/* header */

header{
	margin:			0;
	padding:		0;
}

header div.welcome{
	margin:			0;
	padding:		0;
	height:			50px;
	background:		#DAEFFB;
	position:		relative;
}

header div.welcome p.logo{
	margin:			0;
	padding:		0;
	width:			129px;
	height:			50px;
}

header div.welcome p.logo a{
	display:		block;
	margin:			0;
	padding:		50px 0 0 0;
	width:			129px;
	height:			0;
	overflow:		hidden;
	background:		url(assets/logo.svg) no-repeat 22px 11px;
}


header div.welcome p.welcome{
	margin:			0;
	padding:		0;
	font-size:		14px;
	color:			#003264;
	line-height:		1;
	position:		absolute;
	right:			20px;
	bottom:			8px;
}

header div.welcome p.welcome span{
	font-size:		17px;
}

header div.title{
	margin:			0;
	padding:		0;
	height:			48px;
	background:		#003163;
	position:		relative;
}

header div.title h1{
	margin:			0;
	padding:		0;
	height:			48px;
	position:		absolute;
	left:			100px;
	top:			0;
	font-size:		16px;
	color:			#ffffff;
	line-height:		48px;
	font-weight:		normal;
}

header div.title ul{
	margin:			0;
	padding:		0;
	list-style:		none;
}

header div.title ul li.b1{
	margin:			0;
	padding:		0;
	position:		absolute;
	left:			22px;
	top:			9px;
}

header div.title ul li.b1 a{
	display:		inline-block;
	/display:		inline;
	/zoom:			1;
	margin:			0;
	padding:		0;
	width:			65px;
	height:			30px;
	background:		#ffffff;
	border-radius:		8px;
	font-size:		12px;
	color:			#000000;
	line-height:		30px;
	text-align:		center;
	text-decoration:	none;
}

header div.title ul li.b1 a:hover{
	background:		#CCCCCC;
}

header div.title ul li.b2{
	margin:			0;
	padding:		0;
	position:		absolute;
	right:			22px;
	top:			9px;
}

header div.title ul li.b2 a{
	display:		inline-block;
	/display:		inline;
	/zoom:			1;
	margin:			0;
	padding:		0;
	width:			65px;
	height:			30px;
	background:		#CCD6E0;
	border-radius:		8px;
	font-size:		12px;
	color:			#000000;
	line-height:		30px;
	text-align:		center;
	text-decoration:	none;
}

header div.title ul li.b2 a:hover{
	background:		#B7C5D2;
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
  margin:			0;
  padding:		0;
	display:		block;
	padding:		77px 0 0 135px;
	width:			505px;
	height:			133px;
	font-size:		42px;
	color:			#ffffff;
	line-height:		1.1;
	font-style:		normal;
	text-decoration:	none;
	letter-spacing:		2px;
}

.home article a {
  color: #ffffff;
  text-decoration: none;
}

.home article.reserve{
	width:			960px;
	height:			210px;
	background:	#1C1F84;
}

.home article.inquiry{
	width:			960px;
	height:			210px;
	background:	#00A6ED;
}


.information {
  margin:			0;
  padding:		13px 19px;
  background:		#FFEAB4;
  width:			922px;
}

.information article h2{
	margin:			0 0 8px 0;
	padding:		0 0 8px 0;
	font-size:		18px;
	color:			#000000;
	line-height:		1.1;
	font-weight:		normal;
	border-bottom:		2px dotted #666666;
}

.information article h2 span{
	position:		absolute;
	left:			-5000px;
	top:			-5000px;
}

.information article p{
	margin:			0;
	padding:		0;
	font-size:		13px;
	color:			#666666;
	line-height:		1.4;
}

</style>
