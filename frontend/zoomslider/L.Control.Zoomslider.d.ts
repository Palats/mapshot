// Make sure we are doing module augmentation.
import "leaflet";

declare module "leaflet" {
    interface MapOptions {
        zoomsliderControl?: boolean;
    }
    namespace control {
        export function zoomslider(options: any): any;
    }
}