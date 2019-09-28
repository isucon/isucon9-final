<template>
  <div class="trains">

  <div v-for="(item, index) in reservations">

  <ReservationDetail
    v-bind:reservation="item"
  />
    <div style="clear:both;">
    </div>
    <div class="cancel">
      <button v-on:click="cancelReservation(item.reservation_id)">キャンセル</button>
    </div>
  </div>


  </div>

</template>

<script>
import Router from '@/router.js'
import { apiService } from '../services/api.js'
import ReservationDetail from '@/components/Reservation/ReservationDetail.vue'

export default {
  name: 'reservations',
  components: {ReservationDetail},
  data() {
    return {
      reservations: [],
    }
  },
  methods:{
    getReservations() {
      apiService.getReservations().then((res) => {
        this.reservations = res
      })
    },
    cancelReservation(id) {
      apiService.cancelReservation(id).then((res) => {
        this.getReservations()
      })
    },
  },
  mounted () {
    this.getReservations()
  }
}
</script>

<style scoped>
div.trains {
  background: #18257F;

}

div.trains .condition {
  border-collapse: collapse;
  line-height: 1.1;
  padding: 10px;
}

div.trains .condition div {
  width: 100%;
  text-align: center;
  margin: 3px 0;
}

div.trains .condition .date {
  font-size: 30px;
}

div.trains .condition .station {
  font-size: 28px;
}

div.cancel {
  padding: 15px;
}

div.cancel button {
  border-collapse: separate;
  border-radius: 8px;
  margin-left: auto;
  display:flex;
  justify-content:flex-end;
  align-items:center;
  padding: 5px 10px;
  font-size: 15px;
  text-align: center;
  cursor: pointer;
  outline: none;
  color: #fff;
  background-color: #FF5B5B;
  border: none;
  box-shadow: 0 5px #999;
}

div.cancel button:hover {
  background-color: #FC2A2A;
}

div.cancel button:active {
  background-color: #FC2A2A;
  box-shadow: 0 3px #666;
  transform: translateY(4px);
}
</style>
