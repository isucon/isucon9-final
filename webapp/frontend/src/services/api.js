import { httpService } from './http.js'

class ApiService {

    constructor () {
        this.httpService = httpService
    }

    // 列車検索
    async getTrains (year, month, day) {
        return await this.httpService.get('/trains')
    }

}

export const apiService = new ApiService()
