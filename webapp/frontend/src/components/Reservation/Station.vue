<template>
  <article class="station">
    <select class="from">
      <option>新大阪</option>
      <option>東京</option>
    </select>
    <span>→</span>
    <select class="to">
      <option>新大阪</option>
      <option>東京</option>
    </select>
  </article>
</template>


<script>
export default {
  components: {}
}
</script>

<style>

.station .to,
.station .from {
    width: 300px;
}

.station span {
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
</style>
