<template>
<div v-if="reservation">

  <div class="trains">
    <section class="information">
      <article style="text-align: center;">
        <p style="font-size: 30px; color: red;">
          ご予約はまだ完了していません
        </p>
      </article>
    </section>

    <ReservationDetail
      v-bind:reservation="reservation"
    />

    <section class="information">
      <article>
        <h2>ご案内</h2>
        <p style="font-size: 11px;">
        ■ご利用条件は運送約款の規定によります。<br/>
        </p>
      </article>
    </section>
    </div>

    <div class="card-form">
      <p><label>スーパーセキュアなカードの番号 8桁の数字</label><input v-model="card_number" maxlength="8"/></p>
      <p><label>CVV</label><input v-model="cvv" maxlength="3" size="5"/></p>
      <p>
      <label>有効期限</label>
      <select v-model="expiry_date_month">
        <option>01</option>
        <option>02</option>
        <option>03</option>
        <option>04</option>
        <option>05</option>
        <option>06</option>
        <option>07</option>
        <option>08</option>
        <option>09</option>
        <option>10</option>
        <option>11</option>
        <option>12</option>
      </select> /
      <select v-model="expiry_date_year">
        <option>20</option>
        <option>21</option>
        <option>22</option>
        <option>23</option>
        <option>24</option>
        <option>25</option>
        <option>26</option>
        <option>27</option>
        <option>28</option>
        <option>29</option>
        <option>30</option>
      </select>
      </p>
    </div>

    <div class="button-area">
      <button type="button" class="reserve" v-on:click="payment()">決済する</button>
    </div>
  </div>
</div>




</template>


<script>
import Router from '@/router.js'
import { apiService } from '../services/api.js'
import ReservationDetail from '@/components/Reservation/ReservationDetail.vue'

export default {
  components: {ReservationDetail},
  data: function() {
    return {
      reservation_id: 0,
      reservation: null,
      card_number: "",
      cvv: "",
      expiry_date_month: "01",
      expiry_date_year: "24"
    }
  },
  computed: {
    year() { return this.reservation.date.getYear() + 1900},
    month() { return this.reservation.date.getMonth() + 1 },
    day() { return this.reservation.date.getDate() },
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
    getReservation() {
      apiService.getReservation(this.reservation_id).then((res) => {
        this.reservation = res
        console.log(res)
      })
    },
    payment() {
    var data = {
      card_number: this.card_number,
      cvv: this.cvv,
      expiry_date: this.expiry_date,
    }

    apiService.tokenizeCard(data).then((res) => {
      console.log("OK")
      console.log(res)
      var card_token = res.card_token

      var data = {
        "reservation_id": this.reservation_id,
        "card_token": card_token,
      }
      apiService.commit(data).then((res) => {
        Router.push({path: "/"})
      })
    })

    }
  },
  mounted() {
    this.reservation_id = parseInt(this.$route.query.reservation_id);
    return this.getReservation();
  }
}
</script>

<style scoped>


div.trains {
  background: #18257F;

}

div.trains section.subcontent {
  width: 320px;
  float: left;
  background: #18257F;
  color: #ffffff;
}

div.trains section.trains {
  width: 640px;
  float: right;
  background: #82B1F9;
}

div.trains section.information {
  clear: both;
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

.card-form {

}


.card-form label {
  display: block;
  float: left;
  width: 500px;
  text-align: right;
  margin-right: 30px;
}


.card-form input {
  font-size: 100%;
  text-align: center;
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
