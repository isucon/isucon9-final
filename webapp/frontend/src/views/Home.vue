<template>
  <div>
    <section class="home">

      <article class="reserve">
        <p>
          <router-link to="/reservation">
            <a>予約</a>
          </router-link>
        </p>
      </article>

      <article class="inquiry">
        <p>
          <router-link to="/mypage/reservations">
            <a>予約の確認</a>
          </router-link>
        </p>
      </article>

    </section>

    <section class="information">
      <article class="notice">
        <h2>ご案内
          <span>：　必ずお読みください</span>
        </h2>
        <div class="wrap1">
          <div class="wrap2">
            <p>
              ■ここに何かお知らせをかく<br>
              ・ホゲホゲホゲホゲ<br>
              ・ふがふがふがふが<br>
              ※ピヨピヨ
            </p>
          </div>
        </div>
        <p class="more">
          <span id="l-9"></span>
        </p>
      </article>
    </section>
  </div>

</template>

<script>
export default {
  name: 'home',
  components: {}
}
</script>
