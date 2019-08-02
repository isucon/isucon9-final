<template>
  <article class="person">
    <select class="adult">
      <option>おとな0名</option>
      <option>おとな1名</option>
      <option>おとな2名</option>
      <option>おとな3名</option>
    </select>

    <select class="child">
      <option>こども0名</option>
      <option>こども1名</option>
      <option>こども2名</option>
      <option>こども3名</option>
    </select>
  </article>
</template>


<script>
export default {
  components: {}
}
</script>

<style>
.person .adult,
.person .child {
    width: 320px;
}
</style>
