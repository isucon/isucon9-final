<template>
  <article class="train-item">
    <div class="wrap">
      <div class="header">
          候補 {{ itemIndex }}/{{ itemCount }}
      </div>

      <div class="train">
        <div class="departure">
          <span class="time">{{ item.departure_time.getHours() }}時{{ item.departure_time.getMinutes() }}分 発</span>
          <span class="station">{{ item.departure }}</span>
        </div>

        <div class="name">
          <span class="name">{{ item.train_class }} {{ item.train_name }} 号</span>
          <span class="type">NEKO800系/全席禁煙</span>
        </div>

        <div class="arrival">
          <span class="time">{{ item.arrival_time.getHours() }}時{{ item.arrival_time.getMinutes() }}分 着</span>
          <span class="station">{{ item.arrival }}</span>
        </div>
      </div>

      <div class="available">
        <p class="start" v-on:click="select(item)">この候補を選択</p>
      </div>
    </div>
  </article>
</template>


<script>
export default {
  props: ['itemCount', 'itemIndex', 'item', 'condition', 'select'],
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
