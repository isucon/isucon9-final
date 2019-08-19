import { httpService } from './http.js'
import moment from 'moment';

class ApiService {

    constructor () {
        this.httpService = httpService
    }

    // 列車検索
    async getTrains (condition) {
        var date = new Date(condition.year, condition.month - 1, condition.day)
        console.log(date)
        var params = {
          use_at: moment(date).toISOString(),
          from: condition.from_station,
          to: condition.to_station,
          adult: condition.adult,
          child: condition.child
        }
        return await this.httpService.get('/api/train/search', {"params": params})
    }
}

export const apiService = new ApiService()
