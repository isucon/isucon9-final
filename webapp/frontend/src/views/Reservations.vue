<template>
  <div>

    <section class="reservations">
      <ReservationItem
        v-for="(item, index) in reservations"
        v-bind:item=item
      />
    </section>

  </div>

</template>

<script>
import Router from '@/router.js'
import { apiService } from '../services/api.js'
import ReservationItem from '@/components/Reservations/ReservationItem.vue'

export default {
  name: 'reservations',
  components: {ReservationItem},
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
    }
  },
  mounted () {
    this.getReservations()
  }
}
</script>
