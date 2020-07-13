const http = require('http')
const https = require('https')
const querystring = require('querystring')
const sharp = require('sharp')
const AWS = require('aws-sdk')

const S3 = new AWS.S3({
    signatureVersion: 'v4',
    region: 'us-west-2'
})
const BUCKET = 'cdn.damascus-engineering.com'

exports.handler = async (event, context, callback) => {
    let response = event.Records[0].cf.response

    if (response.status === 404) {
        const request = event.Records[0].cf.request
        const params = querystring.parse(request.querystring)

        if (!params.d) {
            callback(null, response)
            return
        }

        // read the dimension parameter value = width x height and split it by 'x'
        let dimensionMatch = params.d.split('x');

        // read the required path. Ex: uri /alexandria/user/100x100/webp/image.jpg
        let path = request.uri

        // read the S3 key from the path variable.
        // Ex: path variable /alexandria/user/100x100/webp/image.jpg
        let key = path.substring(1)

        // parse the prefix, width, height and image name
        // Ex: key=alexandria/user/200x200/webp/image.jpg
        let prefix, originalKey, match, width, height, requiredFormat, imageName
        let startIndex

        try {
            match = key.match(/(.*)\/(\d+)x(\d+)\/(.*)\/(.*)/)
            prefix = match[1]
            width = parseInt(match[2], 10)
            height = parseInt(match[3], 10)

            // correction for jpg required for 'Sharp'
            requiredFormat = match[4] == 'jpg' ? 'jpeg' : match[4]
            imageName = match[5]
            originalKey = prefix + '/' + imageName
        }
        catch (err) {
            // no prefix exist for image..
            console.log('no prefix present')
            match = key.match(/(\d+)x(\d+)\/(.*)\/(.*)/)
            width = parseInt(match[1], 10)
            height = parseInt(match[2], 10)

            // correction for jpg required for 'Sharp'
            requiredFormat = match[3] == 'jpg' ? 'jpeg' : match[3]
            imageName = match[4]
            originalKey = imageName
        }

        try {
            // get the source image file
            const sourceImg = await S3.getObject({ Bucket: BUCKET, Key: originalKey }).promise()
            
            // perform the resize operation
            const imgBuf = await sharp(sourceImg.Body).resize(width, height).toFormat(requiredFormat).toBuffer()

            const x = await S3.putObject({
                Body: imgBuf,
                Bucket: BUCKET,
                ContentType: 'image/'+ requiredFormat,
                CacheControl: 'max-age=31536000',
                Key: key,
                StorageClass: 'STANDARD'
            }).promise()

            // generate a binary response with resized image
            response.status = 200
            response.body = buffer.toString('base64')
            response.bodyEncoding = 'base64'
            response.headers['content-type'] = [{ 
                key: 'Content-Type', 
                value: "image/" + requiredFormat 
            }]
            callback(null, response)
            return
        } catch (error) {
            // If error, export it to AWS CloudWatch and send default response
            console.log(error)
        }
    }

    callback(null, response)
    return
}
