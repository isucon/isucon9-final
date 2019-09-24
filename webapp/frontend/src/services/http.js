import axios from 'axios'
import Router from '../router.js'

const API_BASE = "";

class ErrorHandler {
    handle (error) {
        if (error.response.status == 401) {
            Router.push({ path: '/login' })
            return;
        }

        if (error.response.data && error.response.data.message ){
          alert(error.response.data.message)
        }

        console.log(error);

        /*
        if (error.response.status === 404) {
            Router.push({ name: 'notfound' })
        } else if (error.response.status === 401) {
            Router.push({ name: 'login' })
        } else if (error.response.status === 500) {
            Router.push({ name: 'internalservererror' })
        }
        */
    }
}

/**
 * Axiosラッパー
 */
export class HttpService {

    constructor (apiBase) {
        const svc = axios.create({
            baseURL: apiBase,
            timeout: 600*1000
        })
        this.svc = svc
    }

    /**
     * APIリクエスト共通処理
     * @param {string} url URL
     * @param {object} config axios設定
     */
    request (url, config = {}) {
        // FIXME: トークン判定処理
        return this.svc.request(url, config).then((response) => {
            // 正常レスポンスハンドリング
            const resp = response.data
            resp.status = response.status
            return resp
        }, (error) => {
            // エラーハンドリング
            const hdl = new ErrorHandler()
            hdl.handle(error)
            return Promise.reject(error)
        })
    }

    /**
     * GETリクエスト送信
     * @param {string} url
     * @param {object} config
     */
    get (url, config = {}) {
        config.method = 'get'
        return this.request(url, config)
    }

    /**
     * POSTリクエスト送信
     * @param {string} url
     * @param {object} data POSTデータ
     * @param {object} config
     */
    post (url, data, config = {}) {
        config.method = 'post'
        config.data = data
        return this.request(url, config)
    }

    /**
     * PUTリクエスト送信
     * @param {string} url
     * @param {object} data PUTデータ
     * @param {object} config
     */
    put (url, data, config = {}) {
        config.method = 'put'
        config.data = data
        return this.request(url, config)
    }

    /**
     * DELETEリクエスト送信
     * @param {string} url
     * @param {object} config
     */
    delete (url, config = {}) {
        config.method = 'delete'
        return this.request(url, config)
    }
}

export const httpService = new HttpService(API_BASE)
