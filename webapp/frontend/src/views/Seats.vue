<template>
<div>

<h2>{{ year }}年{{ month }}月{{ day }}日 {{ train_class }} {{ train_name }}号</h2>
<h2>{{ from_station }} → {{ to_station }}号 {{ car_number }}号車</h2>
<h2>{{ seat_class_name }}　　おとな {{ adult }}名　　こども {{ child }}名 </h2>


<select class="from" v-model="next_car_number" v-on:change="changeCar(next_car_number)">
  <option
    v-for="car in selectableCars"
    v-bind:key="car.car_number"
    v-bind:value="car.car_number"
  >
    {{ car.car_number }}号車
  </option>
</select>

  <div class="seat-select" v-if="seatCols[0]">

    <table v-bind:class="{ filled: filled }">

      <thead>
        <td></td>
        <th
          v-for="(seat, index) in seatCols[0].seats"
          v-bind:key="seat.row"
        >
          {{ seat.row }}
        </th>
      </thead>

      <tr
        v-for="(col, index) in seatCols"
        v-bind:key="col.name"
      >
        <th>{{ col.name }}</th>

        <td
          v-for="(seat, index2) in col.seats"
          v-bind:key="seat.row"
          v-bind:text="seat.text"
          v-bind:seat="seat"
          v-bind:class="{ disabled: seat.disabled , selected: seat.selected}"
          v-on:click="selectSeat(seat)"
        >

          {{ seat.text }}

        </td>

      </tr>

    </table>

  </div>
  <div class="button-area">
    <button type="button" class="reserve" v-on:click="reserve()">予約に進む</button>
  </div>
</div>


</template>


<script>
import Router from '@/router.js'
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
      seat_class: "",
      seatCols: [],
      cars: [],
      next_car_number: null,
    }
  },
  components: {},
  computed: {
    seat_class_name () {
      var m = {
        premium: "プレミアム",
        reserved: "普通席",
        "": "",
      }
      return m[this.seat_class]
    },
    selectableCars () {
      return this.cars.filter(car => car.seat_class == this.seat_class)
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
    },
    selectedSeats () {
      var ret = []
      Object.keys(this.seatCols).forEach(function(key) {
        var col = this[key];
        ret = ret.concat(col.seats.filter(seat => seat.selected));
      }, this.seatCols);
      return ret
    },
    totalCount () {
      return this.adult + this.child
    },
    selectedCount () {
      return this.selectedSeats.length
    },
    availableCount () {
      return this.totalCount - this.selectedCount
    },
    filled () {
      return this.availableCount <= 0
    }
  },
  methods: {
    selectSeat (seat) {
      if (this.availableCount <= 0 && !seat.selected){
        return
      }

      if (seat.text != "○") {
        return
      }

      seat.selected = !seat.selected
      var s = Array.from(this.seatCols)
      this.seatCols = []
      this.seatCols = s
    },
    search() {
      this.seats = []
      this.cars = []
      apiService.getSeats(this.condition).then((res) => {
        this.seats = res["seats"]
        this.cars = res["cars"]

        var self = this
        var ret = []
        var cols = {}

        cols["A"] = {name: "A", seats:[]}
        cols["B"] = {name: "B", seats:[]}
        cols["C"] = {name: "C", seats:[]}
        cols["D"] = {name: "D", seats:[]}
        cols["E"] = {name: "E", seats:[]}

        this.seats.forEach(function(seat){
          seat.text = "○"
          seat.disabled = false
          seat.selected = false

          if(seat.is_occupied){
            seat.text = "×"
            seat.disabled = true

          }
          if(seat.class != self.seat_class){
            seat.text = "-"
            seat.disabled = true
          }

          cols[seat.column]["seats"].push(seat)
        })

        if(cols["A"].seats.length>0)
          ret.push(cols["A"])
        if(cols["B"].seats.length>0)
          ret.push(cols["B"])
        if(cols["C"].seats.length>0)
          ret.push(cols["C"])
        if(cols["D"].seats.length>0)
          ret.push(cols["D"])
        if(cols["E"].seats.length>0)
          ret.push(cols["E"])
        console.log(ret)
        this.seatCols = ret

      })
    },
    changeCar (car_number) {
      console.log(car_number)
      car_number = parseInt(car_number)
      var query = {
        year: this.year,
        month: this.month,
        day: this.day,
        train_class: this.train_class,
        train_name: this.train_name,
        car_number: car_number,
        from_station: this.from_station,
        to_station: this.to_station,
        adult: this.adult,
        child: this.child,
        seat_class: this.seat_class
      }
      Router.push({ path: '/reservation/seats', query: query})
      this.car_number = car_number
      this.search()
    },
    reserve () {
      var condition = {
        year: this.year,
        month: this.month,
        day: this.day,
        train_class: this.train_class,
        train_name: this.train_name,
        car_number: this.car_number,
        from_station: this.from_station,
        to_station: this.to_station,
        adult: this.adult,
        child: this.child,
        seat_class: this.seat_class,
        seats: this.selectedSeats,
        column: "",
      }

      apiService.reserve(condition).then((res) => {
        var query = {
          reservation_id: res.reservation_id,
        }
        Router.push({ path: '/reservation/payment', query: query})
      })
    }
  },
  mounted() {
    this.year = parseInt(this.$route.query.year)
    this.month = parseInt(this.$route.query.month)
    this.day = parseInt(this.$route.query.day)
    this.train_class = this.$route.query.train_class
    this.train_name = this.$route.query.train_name
    this.car_number = parseInt(this.$route.query.car_number)
    this.adult = parseInt(this.$route.query.adult)
    this.child = parseInt(this.$route.query.child)
    this.from_station = this.$route.query.from_station
    this.to_station = this.$route.query.to_station
    this.seat_class = this.$route.query.seat_class
    this.next_car_number = this.car_number
    return this.search();
  }
}
</script>

<style scoped>



.seat-select {
  background-color: #ddd;
}

.seat-select table {
  width: 80%;
  margin-left: auto;
  margin-right: auto;
  text-align: center;
  table-layout: fixed;
}

.seat-select table thead,
.seat-select table tr {
  height: 40px;
}

.seat-select table th,
.seat-select table td {
  border-collapse: collapse;
  border-color: #ddd;
}

.seat-select table th {
  background-color: #0E5AF5;
  color: #ffffff;
}

.seat-select table tr td {
  color: white;
  background-color: #7BA1F9;
  cursor: pointer;
}

.seat-select table.filled tr td {
  background-color: #AAAAAA;
}

.seat-select table tr td.disabled {
  background-color: #777777;
}

.seat-select table tr td.selected {
  background-color: #EB0000;
}


.reserve {
  width: 300px;
  height: 50px;
  line-height: 50px;
  text-align: center;
  margin-left: auto;
  margin-right: auto;
  margin-top: 20px;
  margin-bottom: 20px;

  color: red;
  background: pink;
  font-size: 20px;
  border-width: 1px;
  border-color: #999999;
  border-top-left-radius: 20px;
  border-bottom-left-radius: 20px;
  border-top-right-radius: 20px;
  border-bottom-right-radius: 20px;
  font-weight: bold;
}


.button-area {
  text-align: center;
}

</style>
