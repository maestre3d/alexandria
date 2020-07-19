# Blob Handler Service
The Blob handler module is a serverless service using AWS Lambda, AWS CloudFront, AWS S3 and AWS CloudFormation.
It is responsible of resizing and formatting images using serverless functions triggered by the respective CDN viewer request and origin response.

It uses HTTP communication protocol to receive and respond edge's hits.

Further information can be found [here](https://aws.amazon.com/es/blogs/networking-and-content-delivery/resizing-images-with-amazon-cloudfront-lambdaedge-aws-cdn-blog).

In the following image, you may appreciate the process flow using Amazon Web Services.

![AWS Resizing image architecture](https://d2908q01vomqb2.cloudfront.net/5b384ce32d8cdef02bc3a139d4cac0a22bb029e8/2018/02/20/Social-Media-Image-Resize-Images.png "AWS Resizing architecture")

Alexandria is currently licensed under the MIT license.

## Contribution
Alexandria is an open-source project, that means everyone’s help is appreciated.

If you'd like to contribute, please look at the [Go Contribution Guidelines](https://github.com/maestre3d/alexandria/tree/master/docs/GO_CONTRIBUTION.md).

[Click here](https://github.com/maestre3d/alexandria/tree/master/docs) if you're looking for our docs about engineering, Alexandria API, etc.

## Maintenance
- Main maintainer: [maestre3d](https://github.com/maestre3d)
