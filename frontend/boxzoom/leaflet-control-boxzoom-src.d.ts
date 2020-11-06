// Make sure we are doing module augmentation.
import "leaflet";

declare module "leaflet" {
    namespace Control {
        export function boxzoom(options: any): any;
    }
}