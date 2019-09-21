<template>
<div>
<h2>{{ train_class }} {{ train_name }}号</h2>

<table>

  <tr
  v-for="(seats, index) in seatCols"
  v-bind:seats="seats"
  />


  <td v-for="(seat, index2) in seats" v-bind:seat="seat">
    {{ seat.row }} {{ seat.column }}
  </td>
  </tr>


</table>
</div>



</template>


<script>
import TrainItem from '@/components/Trains/TrainItem.vue'
import { apiService } from '../services/api.js'

export default {
  data: function() {
    return {
      year: null,
      month: null,
      day: null,
      train_class: "",
      train_name: "",
      car_number: 1,
      from_station: null,
      to_station: null,
      adult: null,
      child: null,
      seats: [],
    }
  },
  components: {},
  computed: {
    seatCols() {
      var ret = []

      var cols = {}

      cols["A"] = {name: "A", seats:[]}
      cols["B"] = {name: "A", seats:[]}
      cols["C"] = {name: "A", seats:[]}
      cols["D"] = {name: "A", seats:[]}
      cols["E"] = {name: "A", seats:[]}

      this.seats.forEach(function(seat){
        console.log(seat)
        cols[seat.column]["seats"].push(seat)
      })

      ret.push(cols["A"])
      ret.push(cols["B"])
      ret.push(cols["C"])
      ret.push(cols["D"])
      ret.push(cols["E"])

      return ret
    },
    condition () {
      return {
        year: this.year,
        month: this.month,
        day: this.day,
        train_class: this.train_class,
        train_name: this.train_name,
        car_number: this.car_number,
        from_station: this.from_station,
        to_station: this.to_station,
      }
    }
  },
  methods: {
    search() {
      apiService.getSeats(this.condition).then((res) => {
        this.seats = res["seats"]
      })
    }
  },
  mounted() {
    this.year = this.$route.query.year
    this.month = this.$route.query.month
    this.day = this.$route.query.day
    this.train_class = this.$route.query.train_class
    this.train_name = this.$route.query.train_name
    this.car_number = this.$route.query.car_number
    this.adult = this.$route.query.adult
    this.child = this.$route.query.child
    this.from_station = this.$route.query.from_station
    this.to_station = this.$route.query.to_station
    return this.search();
  }
}
</script>
