<template>
  <article class="reservation-item">
    <div class="wrap">
      <div class="header">
          予約ID {{ item.reservation_id }} {{ item.date }}
      </div>

      <div class="train">
        <div class="departure">
          <span class="time">{{ item.departure_time }} 発</span>
          <span class="station">{{ item.departure }}</span>
        </div>

        <div class="name">
          <span class="name">{{ item.train_class }} {{ item.train_name }} 号</span>
          <span class="type">NEKO800系/全席禁煙</span>
        </div>

        <div class="arrival">
          <span class="time">{{ item.arrival_time }} 着</span>
          <span class="station">{{ item.arrival }}</span>
        </div>
      </div>

    </div>
  </article>
</template>


<script>
export default {
  props: ['item',],
  data: function() {
    return {
    }
  },
  components: {},
  methods: {
  }
}
</script>

<style scoped>

</style>
