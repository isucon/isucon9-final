<template>
  <div>
    <section class="subcontent">
      <article class="condition">
        <div class="date">{{ year }}年{{ month }}月{{ day }}日</div>
        <div class="station">{{ reservation.departure }}→{{ reservation.arrival }}</div>
        <div class="person">おとな {{ reservation.adult }} 名 こども {{ reservation.child }} 名</div>
      </article>
    </section>



    <section class="trains">
      <article class="train-item">
        <div class="wrap">

          <div class="train">
            <div class="departure">
              <span class="time">{{ departure_time.getHours() }}時{{ departure_time.getMinutes() }}分 発</span>
              <span class="station">{{ reservation.departure }}</span>
            </div>

            <div class="name">
              <span class="name">{{ reservation.train_class }} {{ reservation.train_name }} 号</span>
              <span class="type">NEKO800系/全席禁煙</span>
            </div>

            <div class="arrival">
              <span class="time">{{ arrival_time.getHours() }}時{{ arrival_time.getMinutes() }}分 着</span>
              <span class="station">{{ reservation.arrival }}</span>
            </div>

            <div class="seats">

              <h3>{{ seat_class_name }}</h3>


              <div v-if="reservation.seat_class != 'non-reserved'">
                <p v-for="(seat, index) in reservation.seats" style="margin-top: 0; margin-bottom: 0;">
                  {{ reservation.car_number }}号車{{ seat.seat_row }}番{{ seat.seat_column }}席
                </p>
              </div>

            </div>

          </div>
        </div>
      </article>
    </section>
    <section class="price">
      <article class="price-item">
        <div class="wrap">
          <p>おとな</p><p style="text-align: right; margin-top: -30px;">{{ reservation.adult }}名分</p>
          <p>こども</p><p style="text-align: right; margin-top: -30px;">{{ reservation.child }}名分</p>
          <p>合計</p><p style="text-align: right; margin-top: -30px;">¥{{ reservation.amount }}</p>
        </div>
      </article>
    </section>
  </div>
</template>


<script>
export default {
  props: ['reservation',],
  data: function() {
    return {
    }
  },
  components: {},
  computed: {
    year() { return new Date(this.reservation.date).getYear() + 1900},
    month() { return new Date(this.reservation.date).getMonth() + 1 },
    day() { return new Date(this.reservation.date).getDate() },
    expiry_date() { return this.expiry_date_month + "/" + this.expiry_date_year },
    arrival_time() {
      return new Date("2020-01-01 " + this.reservation.arrival_time)
    },
    departure_time() {
      return new Date("2020-01-01 " + this.reservation.departure_time)
    },
    seat_class_name () {
      var m = {
        premium: "プレミアム",
        reserved: "普通席",
        "non-reserved": "自由席",
        "": "",
      }
      return m[this.reservation.seat_class]
    },
  },
  methods: {
  }
}
</script>

<style scoped>

</style>
