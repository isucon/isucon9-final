<template>
  <div>
    <section class="home">

      <article class="reserve">
        <p><a>予約</a></p>
      </article>

      <article class="inquiry">
        <p><a>予約の確認</a></p>
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
// @ is an alias to /src
import HelloWorld from '@/components/HelloWorld.vue'

export default {
  name: 'home',
  components: {
    HelloWorld
  }
}
</script>
